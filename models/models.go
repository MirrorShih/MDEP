package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Detector struct {
	Id     primitive.ObjectID `json:"detector_id" bson:"_id,omitempty"`
	Name   string             `json:"detector_name" bson:"name,omitempty"`
	FileId primitive.ObjectID `json:"file_id" bson:"file_id,omitempty"`
}

type Report struct {
	Id             primitive.ObjectID `json:"report_id" bson:"_id,omitempty"`
	FuncType       string             `json:"function_type" bson:"function_type,omitempty"`
	Accuracy       float64            `json:"accuracy" bson:"accuracy,omitempty"`
	FP             float64            `json:"fp" bson:"fp,omitempty"`
	FN             float64            `json:"fn" bson:"fn,omitempty"`
	Precision      float64            `json:"precision" bson:"precision,omitempty"`
	Recall         float64            `json:"recall" bson:"recall,omitempty"`
	F1             float64            `json:"f1" bson:"f1,omitempty"`
	AvgTime        float64            `json:"avg_time" bson:"avg_time,omitempty"`
	MinTime        float64            `json:"min_time" bson:"min_time,omitempty"`
	MaxTime        float64            `json:"max_time" bson:"max_time,omitempty"`
	TestingTime    float64            `json:"testing_time" bson:"testing_time,omitempty"`
	TestSampleNum  float64            `json:"testing_sample_num" bson:"testing_sample_num,omitempty"`
	TotalSampleNum float64            `json:"total_sample_num" bson:"total_sample_num,omitempty"`
}

type GitHubUser struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTML_URL          string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
	Name              string `json:"name"`
	Company           string `json:"company"`
	Blog              string `json:"blog"`
	Location          string `json:"location"`
	Email             string `json:"email"`
	Hireable          bool   `json:"hireable"`
	Bio               string `json:"bio"`
	PublicRepos       int    `json:"public_repos"`
	PublicGists       int    `json:"public_gits"`
	Followers         int    `json:"followers"`
	Following         int    `json:"following"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type Task struct {
	Id primitive.ObjectID `json:"report_id" bson:"_id,omitempty"`
}

type Dataset struct {
	Name string `json:"dataset_name" bson:"dataset_name,omitempty"`
}
