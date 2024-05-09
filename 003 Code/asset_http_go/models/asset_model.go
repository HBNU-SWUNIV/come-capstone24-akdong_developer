package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Asset struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name          string             `bson:"name" json:"name" validate:"required"`
	CategoryID    int                `bson:"categoryid" json:"categoryid" validate:"required"`
	Thumbnail     string             `bson:"thumbnail" json:"thumbnail" validate:"required"` // 썸네일의
	ThumbnailExt  string             `bson:"thumbnailext" json:"thumbnailext" validate:"required"`
	File          string             `bson:"file" json:"file" validate:"required"` // 파일 DB의 PK
	UploadDate    string             `bson:"uploaddate" json:"uploaddate"`
	DownloadCount int                `bson:"downloadcount" json:"downloadcount"`
	Price         int                `bson:"price" json:"price"`
	IsDisable     bool               `bson:"isdisable" json:"isdisable"`
}

type SearchResult struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Thumbnail    string `json:"thumbnail"`
	ThumbnailExt string `json:"thumbnailext"`
}

type DownFile struct {
	ID   string `json:"id"`
	File string `json:"file"`
}
