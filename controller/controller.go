package controller

import (
	"MDEP/models"
	"MDEP/services"
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/Jeffail/tunny"
	"golang.org/x/oauth2"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var pool *tunny.Pool

type DetectorTask struct {
	id        primitive.ObjectID
	taskPath  string
	functions []string
	reportId  []string
}

func init() {
	numCPUs := runtime.NumCPU()
	pool = tunny.NewFunc(numCPUs, func(task interface{}) interface{} {
		detectorTask := task.(DetectorTask)
		RunTask(detectorTask.id, detectorTask.taskPath, detectorTask.functions, detectorTask.reportId)
		return true
	})
	pool.SetSize(1)
}

func HandleCallback(c *gin.Context) {
	ctx := context.Background()
	ac := oauth2.Config{
		ClientID:     os.Getenv("GITHUB_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_OAUTH_CLIENT_SECRECT"),
		RedirectURL:  os.Getenv("GITHUB_OAUTH_REDIRECT_URL"),
		Scopes:       []string{"read:user", "repo"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}

	code := c.Query("code")
	token, err := ac.Exchange(ctx, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Use the token to make authenticated requests to GitHub API
	client := ac.Client(ctx, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer resp.Body.Close()
	var githubUser models.GitHubUser
	// parse the JSON response into the 'GitHubUser' struct.
	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Set("github_user", githubUser)

	c.JSON(http.StatusOK, gin.H{
		"user_data: ": githubUser,
	})
}

func GetDetectorList(c *gin.Context) {
	results := services.MongoClient.ListDetector("MDEP", "detector")
	var response []models.Detector
	for _, result := range results {
		response = append(response, result)
	}
	c.JSON(http.StatusOK, response)
}

func GetDetector(c *gin.Context) {
	target := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(target)
	filter := bson.D{bson.E{Key: "_id", Value: id}}
	results := services.MongoClient.GetCertainDetector("MDEP", "detector", filter)
	c.JSON(http.StatusOK, results)
}

func CreateDetector(c *gin.Context) {
	uploadFile, _ := c.FormFile("file")
	log.Println(uploadFile.Filename)
	uploadFolder := "/home/MDEP/upload/"
	uploadFilePath := uploadFolder + uploadFile.Filename
	defer os.Remove(uploadFilePath)
	c.SaveUploadedFile(uploadFile, uploadFilePath)
	services.MongoClient.InsertDetector("MDEP", uploadFilePath, uploadFile.Filename, "detector")
	c.JSON(http.StatusNoContent, gin.H{})
}

type TaskRequest struct {
	DetectorId   string   `json:"detector_id"`
	FunctionType []string `json:"function_type"`
}

func DownloadDetector(detectorId primitive.ObjectID, taskPath string) {
	filter := bson.D{bson.E{Key: "_id", Value: detectorId}}
	results := services.MongoClient.GetCertainDetector("MDEP", "detector", filter)
	services.MongoClient.DownloadFile("MDEP", results.FileId, taskPath+"detector.zip")
}

func InitDetector(taskPath string) {
	cmd := exec.Command("unzip", taskPath+"detector.zip", "-d", taskPath)
	err := cmd.Run()
	if err != nil {
		log.Printf("cmd.Run() failed with %s\n", err)
	}
	cmd = exec.Command("python", "scripts/envBuilder.py", taskPath+"taskEnv")
	err = cmd.Run()
	if err != nil {
		log.Printf("cmd.Run() failed with %s\n", err)
	}
	cmd = exec.Command(taskPath+"taskEnv/bin/pip", "install", "-r", taskPath+"requirements.txt")
	err = cmd.Run()
	if err != nil {
		log.Printf("cmd.Run() failed with %s\n", err)
	}
	cmd = exec.Command("cp", "/src/scripts/metrics.py", taskPath)
	err = cmd.Run()
	if err != nil {
		log.Printf("cmd.Run() failed with %s\n", err)
	}
}

func RunDetector(taskPath string, functions, reportId []string) {
	datasetPath := "/mnt/dataset/"
	for i, function := range functions {
		startTime := time.Now()
		minTime := -1.
		maxTime := -1.
		totalNum := 0.0
		err := filepath.Walk(datasetPath+function, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if info.Name() == "dataset.csv" {
				return nil
			}
			totalNum += 1
			startFileTime := time.Now()
			cmd := exec.Command("taskEnv/bin/python", "main.py", "-i", path, "-o", function+"records.csv", "-c")
			cmd.Dir = taskPath
			err = cmd.Run()
			if err != nil {
				log.Printf("cmd.Run() failed with %s\n", err)
				return err
			}
			executionTime := time.Since(startFileTime).Seconds()
			if minTime == -1 {
				minTime = executionTime
				maxTime = executionTime
			} else {
				minTime = math.Min(minTime, executionTime)
				maxTime = math.Max(maxTime, executionTime)
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
		testingTime := time.Since(startTime).Seconds()
		cmd := exec.Command("python", taskPath+"metrics.py", function)
		cmd.Dir = taskPath
		err = cmd.Run()
		if err != nil {
			log.Printf("cmd.Run() failed with %s\n", err)
		}
		reportID, err := primitive.ObjectIDFromHex(reportId[i])
		if err != nil {
			log.Printf("cannot covert report id")
		}
		content, err := os.OpenFile(taskPath+"metrics.csv", os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Println("Cannot find csv file:", taskPath+"metrics.csv", err)
		}
		r := csv.NewReader(content)
		title := true
		var accuracy, testSampleNum, precision, recall, f1_score float64
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Println(err)
			}
			if title {
				title = false
				continue
			}
			testSampleNum, err = strconv.ParseFloat(record[0], 64)
			if err != nil {
				log.Println(err)
			}
			accuracy, err = strconv.ParseFloat(record[1], 64)
			if err != nil {
				log.Println(err)
			}
			precision, err = strconv.ParseFloat(record[2], 64)
			if err != nil {
				log.Println(err)
			}
			recall, err = strconv.ParseFloat(record[3], 64)
			if err != nil {
				log.Println(err)
			}
			f1_score, err = strconv.ParseFloat(record[4], 64)
			if err != nil {
				log.Println(err)
			}
		}
		services.MongoClient.InsertReport("MDEP", "report", models.Report{reportID, function, accuracy, 0, 0, precision, recall, f1_score, testingTime / totalNum, minTime, maxTime, testingTime, testSampleNum, totalNum})
	}
}

func RunTask(detectorId primitive.ObjectID, taskPath string, functions, reportId []string) {
	DownloadDetector(detectorId, taskPath)
	InitDetector(taskPath)
	RunDetector(taskPath, functions, reportId)
	os.RemoveAll(taskPath)
}

func CreateTask(c *gin.Context) {
	var json TaskRequest
	c.BindJSON(&json)
	id, _ := primitive.ObjectIDFromHex(json.DetectorId)
	log.Println(json.DetectorId)
	log.Printf("%v", &json)
	taskPath := "/home/MDEP/task/"
	os.Mkdir(taskPath, os.ModePerm)
	var reportID []string
	for _ = range json.FunctionType {
		reportID = append(reportID, primitive.NewObjectID().Hex())
	}
	go pool.Process(DetectorTask{id, taskPath, json.FunctionType, reportID})
	c.JSON(http.StatusOK, gin.H{"report_id": reportID})
}

func GetReportList(c *gin.Context) {
	results := services.MongoClient.ListReport("MDEP", "report")
	var response []models.Report
	for _, result := range results {
		response = append(response, result)
	}
	c.JSON(http.StatusOK, response)
}

func GetReport(c *gin.Context) {
	target := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(target)
	filter := bson.D{bson.E{Key: "_id", Value: id}}
	result := services.MongoClient.GetCertainReport("MDEP", "report", filter)
	c.JSON(http.StatusOK, result)
}

func UpdateDetector(c *gin.Context) {
	// TODO
	target := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(target)
	filter := bson.D{bson.E{Key: "_id", Value: id}}
	if services.MongoClient.PatchDetector("MDEP", "detector", filter) {
		result := services.MongoClient.GetCertainDetector("MDEP", "detector", filter)
		c.JSON(http.StatusOK, result)
	}
}

func DeleteDetector(c *gin.Context) {
	target := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(target)
	filter := bson.D{bson.E{Key: "_id", Value: id}}
	result := services.MongoClient.DeleteDetector("MDEP", "detector", filter)
	if result == true {
		c.JSON(http.StatusOK, gin.H{})
	} else {
		log.Println("delete failed")
	}
}
