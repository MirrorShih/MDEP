package controller

import (
	"MDEP/models"
	"MDEP/services"
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type AuthController struct {
	oauthConfig *oauth2.Config
}

func NewAuthController(clientID, clientSecret string) *AuthController {
	return &AuthController{
		oauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
			RedirectURL:  os.Getenv("GITHUB_OAUTH_REDIRECT_URL"),
		},
	}
}

func (ac *AuthController) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "", gin.H{})
}

func (ac *AuthController) InitiateGitHubOAuth(c *gin.Context) {
	authURL := ac.oauthConfig.AuthCodeURL("state")
	c.Redirect(http.StatusFound, authURL)
}

func (ac *AuthController) HandleGitHubCallback(c *gin.Context) {
	// get authorization code from redirect URL
	code := c.Query("code")
	// change authorization code into access token
	token, err := ac.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Use the token to make authenticated requests to GitHub API
	client := ac.oauthConfig.Client(context.Background(), token)
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

	c.JSON(http.StatusOK, gin.H{
		"access_token": token.AccessToken,
		"user_data":    githubUser,
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
	DetectorId   string `json:"detector_id"`
	FunctionType string `json:"function_type"`
}

func CreateTask(c *gin.Context) {
	// TODO
	var json TaskRequest
	c.BindJSON(&json)
	id, _ := primitive.ObjectIDFromHex(json.DetectorId)
	log.Println(json.DetectorId)
	log.Printf("%v", &json)
	filter := bson.D{bson.E{Key: "_id", Value: id}}
	results := services.MongoClient.GetCertainDetector("MDEP", "detector", filter)
	taskPath := "/home/MDEP/task/"
	os.Mkdir(taskPath, os.ModePerm)
	defer os.RemoveAll(taskPath)
	services.MongoClient.DownloadFile("MDEP", results.FileId, taskPath+"detector.zip")
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
	cmd = exec.Command("taskEnv/bin/python", "main.py", "-i", "TestingBin/0000dc2f3c8bde2d3b61cd1ba3aa5e839c0a7bf432d2e06a88a7ce3b199453e7", "-o", "myDetector_FC_records.csv", "-c")
	cmd.Dir = taskPath
	err = cmd.Run()
	if err != nil {
		log.Printf("cmd.Run() failed with %s\n", err)
	}
	content, err := os.OpenFile(taskPath+"myDetector_FC_records.csv", os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatalln("Cannot find csv file:", taskPath+"myDetector_FC_records.csv", err)
	}
	r := csv.NewReader(content)
	r.Comma = ','
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("%v\n", record)
	}
	c.JSON(http.StatusNoContent, gin.H{})
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
