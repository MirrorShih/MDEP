package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type DetectorRes struct {
	Id     primitive.ObjectID `json:"detector_id" bson:"_id,omitempty"`
	Name   string             `json:"detector_name" bson:"name,omitempty"`
	FileId primitive.ObjectID `json:"file_id" bson:"file_id,omitempty"`
}

type ReportRes struct {
	Id             primitive.ObjectID `json:"report_id" bson:"_id,omitempty"`
	FuncType       string             `json:"function_type" bson:"function_type,omitempty"`
	Accuracy       float64            `json:"accuracy" bson:"accuracy,omitempty"`
	FP             float64            `json:"fp" bson:"fp,omitempty"`
	FN             float64            `json:"fn" bson:"fn,omitempty"`
	Precision      float64            `json:"precision" bson:"precision,omitempty"`
	Recall         float64            `json:"recall" bson:"recall,omitempty"`
	F1             float64            `json:"f1" bson:"f1,omitempty"`
	TestTime       float64            `json:"testing_time" bson:"testing_time,omitempty"`
	TestSampleNum  float64            `json:"testing_sample_num" bson:"testing_sample_num,omitempty"`
	TotalSampleNum float64            `json:"total_sample_num" bson:"total_sample_num,omitempty"`
}
