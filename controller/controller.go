package controller

import (
	"MDEP/models"
	"MDEP/services"
	"context"
	"encoding/csv"
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
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var pool *tunny.Pool

type DetectorTask struct {
	detector  *models.Detector
	taskPath  string
	functions []string
	reportId  []models.Task
}

func init() {
	numCPUs := runtime.NumCPU()
	pool = tunny.NewFunc(numCPUs, func(task interface{}) interface{} {
		detectorTask := task.(DetectorTask)
		RunTask(detectorTask.detector, detectorTask.taskPath, detectorTask.functions, detectorTask.reportId)
		return true
	})
	pool.SetSize(1)
}

var (
	AuthController = oauth2.Config{
		ClientID:     os.Getenv("GITHUB_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_OAUTH_CLIENT_SECRECT"),
		RedirectURL:  os.Getenv("GITHUB_OAUTH_REDIRECT_URL"),
		Scopes:       []string{"read:user", "repo"},
		Endpoint:     github.Endpoint,
	}
	Store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
)

func HandleCallback(c *gin.Context) {
	ctx := context.Background()

	code := c.Query("code")
	token, err := AuthController.Exchange(ctx, code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"token exchange failed": err.Error()})
		return
	}

	session, _ := Store.Get(c.Request, "session1")
	session.Values["access_token"] = token.AccessToken
	session.Save(c.Request, c.Writer)

	c.Redirect(http.StatusFound, "http://140.118.155.18:8001/dash")
}

func UserFilter(c *gin.Context) (int, string) {
	githubUser, _ := c.Get("github_user")
	userID := githubUser.(models.GitHubUser).ID
	userName := githubUser.(models.GitHubUser).Name
	return userID, userName
}

func GetDetectorList(c *gin.Context) {
	userID, _ := UserFilter(c)
	results := services.MongoClient.ListDetector("MDEP", "detector", userID)
	var response []models.Detector
	for _, result := range results {
		response = append(response, result)
	}
	c.JSON(http.StatusOK, response)
}

func GetDetector(c *gin.Context) {
	userID, _ := UserFilter(c)
	target := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(target)
	filter := bson.D{bson.E{Key: "_id", Value: id}}
	filter = append(filter, bson.E{Key: "user_id", Value: userID})
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
	userID, userName := UserFilter(c)
	services.MongoClient.InsertDetector("MDEP", uploadFilePath, uploadFile.Filename, "detector", userID, userName)
	c.JSON(http.StatusNoContent, gin.H{})
}

type TaskRequest struct {
	DetectorId   string   `json:"detector_id"`
	FunctionType []string `json:"function_type"`
}

type DescriptionRequest struct {
	Content string `json:"content"`
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

func RunDetector(taskPath string, functions []string, reportId []models.Task, detector *models.Detector) {
	datasetPath := "/mnt/dataset/"
	for i, function := range functions {
		startTime := time.Now()
		minTime := -1.
		maxTime := -1.
		totalNum := 0.0
		err := filepath.Walk(datasetPath+function, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				totalNum += 1
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
		reportID, err := primitive.ObjectIDFromHex(reportId[i].Id)
		if err != nil {
			log.Printf("cannot covert report id")
		}
		userID := reportId[i].UserID
		userName := reportId[i].UserName

		content, err := os.OpenFile(taskPath+"metrics.csv", os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Println("Cannot find csv file:", taskPath+"metrics.csv", err)
			services.MongoClient.InsertReport("MDEP", "report", models.Report{reportID, function, -1, -1, -1, -1, -1, -1, testingTime / totalNum, minTime, maxTime, testingTime, -1, totalNum, userID, userName, primitive.NewDateTimeFromTime(time.Now()), detector.Name})
		} else {
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
			services.MongoClient.InsertReport("MDEP", "report", models.Report{reportID, function, accuracy, 0, 0, precision, recall, f1_score, testingTime / totalNum, minTime, maxTime, testingTime, testSampleNum, totalNum, userID, userName, primitive.NewDateTimeFromTime(time.Now()), detector.Name})
		}
	}
}

func RunTask(detector *models.Detector, taskPath string, functions []string, reportId []models.Task) {
	DownloadDetector(detector.Id, taskPath)
	InitDetector(taskPath)
	RunDetector(taskPath, functions, reportId, detector)
	os.RemoveAll(taskPath)
}

func CreateTask(c *gin.Context) {
	var json TaskRequest
	c.BindJSON(&json)
	id, _ := primitive.ObjectIDFromHex(json.DetectorId)
	filter := bson.D{bson.E{Key: "_id", Value: id}}
	detector := services.MongoClient.GetCertainDetector("MDEP", "detector", filter)
	log.Println(json.DetectorId)
	log.Printf("%v", &json)
	taskPath := "/home/MDEP/task/"
	os.Mkdir(taskPath, os.ModePerm)
	var reportID []models.Task
	userID, userName := UserFilter(c)
	for _ = range json.FunctionType {
		reportID = append(reportID, models.Task{Id: primitive.NewObjectID().Hex(), UserID: userID, UserName: userName})
	}
	go pool.Process(DetectorTask{detector, taskPath, json.FunctionType, reportID})
	c.JSON(http.StatusOK, reportID)
}

func GetReportList(c *gin.Context) {
	userID, _ := UserFilter(c)
	results := services.MongoClient.ListReport("MDEP", "report", userID)
	var response []models.Report
	for _, result := range results {
		response = append(response, result)
	}
	c.JSON(http.StatusOK, response)
}

func GetReport(c *gin.Context) {
	userID, _ := UserFilter(c)
	target := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(target)
	filter := bson.D{bson.E{Key: "_id", Value: id}}
	filter = append(filter, bson.E{Key: "user_id", Value: userID})
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

func GetDatasetList(c *gin.Context) {
	files, err := os.ReadDir("/mnt/dataset")
	if err != nil {
		log.Fatal(err)
	}
	var response []models.Dataset
	for _, result := range files {
		if result.IsDir() {
			response = append(response, models.Dataset{Name: result.Name()})
		}
	}
	c.JSON(http.StatusOK, response)
}

func GetLeaderboard(c *gin.Context) {
	target := c.Param("dataset")
	filter := bson.D{
		{"$and",
			bson.A{
				bson.D{{"function_type", bson.D{{"$eq", target}}}},
				bson.D{{"accuracy", bson.D{{"$ne", -1}}}},
			},
		},
	}
	results := services.MongoClient.ListLeaderboard("MDEP", "report", filter)
	var response []models.Report
	for _, result := range results {
		response = append(response, result)
	}
	c.JSON(http.StatusOK, response)
}

func DeleteReport(c *gin.Context) {
	userID, _ := UserFilter(c)
	target := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(target)
	filter := bson.D{bson.E{Key: "_id", Value: id}}
	filter = append(filter, bson.E{Key: "user_id", Value: userID})
	result := services.MongoClient.DeleteReport("MDEP", "report", filter)
	if result == true {
		c.JSON(http.StatusOK, gin.H{})
	} else {
		log.Println("delete failed")
		c.JSON(http.StatusUnauthorized, gin.H{})
	}
}

func UpdateDescription(c *gin.Context) {
	userID, _ := UserFilter(c)
	var json DescriptionRequest
	c.BindJSON(&json)
	target := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(target)
	filter := bson.D{bson.E{Key: "_id", Value: id}}
	filter = append(filter, bson.E{Key: "user_id", Value: userID})
	update := bson.D{{"$set", bson.D{{"description", json.Content}}}}
	result := services.MongoClient.UpdateDescription("MDEP", "detector", filter, update)
	if result == true {
		c.JSON(http.StatusNoContent, gin.H{})
	} else {
		log.Println("update failed")
		c.JSON(http.StatusUnauthorized, gin.H{})
	}
}
