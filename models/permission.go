package models

type Permission struct {
	Subject string `bson:"subject" json:"subject" binding:"required"`
	Action  string `bson:"action" json:"action" binding:"required"`
	Object  string `bson:"object" json:"object" binding:"required"`
}

type GroupingPolicy struct {
	User string `bson:"user" json:"user" binding:"required"`
	Role string `bson:"role" json:"role" binding:"required"`
}
