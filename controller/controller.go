package controller

import (
	"MDEP/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"os"
)

func GetDetectorList(c *gin.Context) {
	results := services.MongoClient.ListDetector("MDEP", "detector")
	var response []bson.M
	for _, result := range results {
		response = append(response, bson.M{"detector_id": result.Id.Hex(), "detector_name": result.Name})
	}
	c.JSON(http.StatusOK, response)
}

func GetDetector(c *gin.Context) {

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

func CreateTask(c *gin.Context) {
	c.JSON(http.StatusNoContent, gin.H{})
}

func GetReportList(c *gin.Context) {
	results := services.MongoClient.ListReport("MDEP", "report")
	var response []bson.M
	for _, result := range results {
		response = append(response, bson.M{"report_id": result.Id.Hex(), "function_type": result.FuncType,
			"accuracy": result.Accuracy, "false_positive": result.FP,
			"false_negative": result.FN, "precision": result.Precision,
			"recall": result.Recall, "f1_score": result.F1,
			"testing_time": result.TestTime, "testing_sample_num": result.TestSampleNum,
			"total_sample_num": result.TotalSampleNum})
	}
	c.JSON(http.StatusOK, response)
}

func GetReport(c *gin.Context) {
	target := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(target)
	filter := bson.D{bson.E{Key: "_id", Value: id}}
	result := services.MongoClient.GetCertainReport("MDEP", "report", filter)
	c.JSON(http.StatusOK, gin.H{"report_id": result.Id.Hex(), "function_type": result.FuncType,
		"accuracy": result.Accuracy, "false_positive": result.FP,
		"false_negative": result.FN, "precision": result.Precision,
		"recall": result.Recall, "f1_score": result.F1,
		"testing_time": result.TestTime, "testing_sample_num": result.TestSampleNum,
		"total_sample_num": result.TotalSampleNum})
}

func UpdateDetector(c *gin.Context) {
	target := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(target)
	filter := bson.D{bson.E{Key: "_id", Value: id}}
	if services.MongoClient.PatchDetector("MDEP", "detector", filter) {
		result := services.MongoClient.GetCertainDetector("MDEP", "detector", filter)
		c.JSON(http.StatusOK, gin.H{"detector_id": result.Id.Hex(), "detector_name": result.Name})
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
