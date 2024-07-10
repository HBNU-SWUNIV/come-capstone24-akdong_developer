package controllers

import (
	"asset_http_go/configs"
	"asset_http_go/models"
	"asset_http_go/responses"
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var assetCollection *mongo.Collection = configs.GetCollection(configs.DB, "assets")
var validate = validator.New()

func CreateAsset() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var asset models.Asset
		defer cancel()

		//validate the request body
		if err := c.BindJSON(&asset); err != nil {
			c.JSON(http.StatusBadRequest, responses.AssetResponse{Code: 0, Message: "Request validation failed: " + err.Error()})
			log.Println("(create)BindJsonErr :", err)
			return
		}

		//use the validator library to validate required fields
		if validationErr := validate.Struct(&asset); validationErr != nil {
			c.JSON(http.StatusBadRequest, responses.AssetResponse{Code: 0, Message: "Validation error: " + validationErr.Error()})
			log.Println("(create)BindJsonValidation error :")
			return
		}

		newAsset := models.Asset{
			ID:            primitive.NewObjectID(),
			Name:          asset.Name,
			CategoryID:    asset.CategoryID,
			Thumbnail:     asset.Thumbnail,
			ThumbnailExt:  asset.ThumbnailExt,
			File:          asset.File,
			UploadDate:    time.Now(),
			DownloadCount: asset.DownloadCount,
			Price:         asset.Price,
			IsDisable:     asset.IsDisable,
		}

		// result, err := assetCollection.InsertOne(ctx, newAsset)
		if _, err := assetCollection.InsertOne(ctx, newAsset); err != nil {
			log.Println("Database insertion error :", err)
			c.JSON(http.StatusInternalServerError, responses.AssetResponse{Code: 0, Message: "Database insertion error: " + err.Error()})
			return
		}

		log.Println("Create New Asset :", newAsset)
		//c.JSON(http.StatusCreated, responses.AssetResponse{Code: 1, Message: "success"})
		c.JSON(http.StatusCreated, responses.AssetResponse{Code: 1, Message: "Asset created successfully"})
		// c.JSON(http.StatusCreated, responses.AssetResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"data": result}})
	}
}

func GetAsset() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var results []models.AssetInfo
		// asset id
		id := c.Query("id")

		filter := bson.M{}
		if id != "" {
			objID, err := primitive.ObjectIDFromHex(id) // 문자열 ID를 ObjectID로 변환
			if err != nil {
				log.Println("Invalid ID format :", err)
				c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "Invalid ID format"})
				return
			}
			filter["_id"] = objID // 정확한 ObjectID로 필터링
		}

		cur, err := assetCollection.Find(ctx, filter)
		if err != nil {
			log.Println("Database query failed :", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "Database query failed"})
			return
		}
		defer cur.Close(ctx)

		for cur.Next(ctx) {
			var asset models.Asset
			if err := cur.Decode(&asset); err != nil {
				continue
			}
			result := models.AssetInfo{
				ID:            asset.ID.Hex(),
				Name:          asset.Name,
				CategoryID:    strconv.Itoa(asset.CategoryID), // int를 string으로 변환
				Thumbnail:     asset.Thumbnail,
				ThumbnailExt:  asset.ThumbnailExt,
				File:          asset.File,
				UploadDate:    asset.UploadDate.Format(time.RFC3339), // time.Time을 RFC3339 문자열로 변환
				DownloadCount: strconv.Itoa(asset.DownloadCount),     // int를 string으로 변환
				Price:         strconv.Itoa(asset.Price),             // int를 string으로 변환
				IsDisable:     strconv.FormatBool(asset.IsDisable),   // bool을 string으로 변환
			}
			results = append(results, result)
		}

		if err := cur.Err(); err != nil {
			log.Println("Error reading from database :", err)
			c.JSON(http.StatusInternalServerError, responses.AssetResponse{Code: 0, Message: "Error reading from database: " + err.Error()})
			return
		}

		if len(results) == 0 {
			log.Println("Success, but no assets found matching :", results)
			c.JSON(http.StatusOK, responses.AssetResponse{Code: 1, Message: "Success, but no assets found matching the criteria", Data: results})
		} else {
			log.Println("Success, asset_info :", results)
			c.JSON(http.StatusOK, responses.AssetResponse{Code: 1, Message: "Assets retrieved successfully", Data: results})
		}
	}
}

func SearchAsset() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var results []models.SearchResult

		name := c.Query("name")
		categoryidStr := c.Query("categoryid")

		filter := bson.M{}
		if name != "" {
			filter["name"] = bson.M{"$regex": primitive.Regex{Pattern: name, Options: "i"}} // 대소문자 구분 없이 이름 검색
		}
		if categoryidStr != "" {
			categoryid, err := strconv.Atoi(categoryidStr) // 문자열을 정수로 변환
			if err != nil {
				log.Println("Invalid category ID :", err)
				c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "Invalid category ID"})
				return
			}
			filter["categoryid"] = categoryid
		}

		cur, err := assetCollection.Find(ctx, filter)
		if err != nil {
			log.Println("Database query failed :", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "Database query failed"})
			return
		}
		defer cur.Close(ctx)

		for cur.Next(ctx) {
			var asset models.Asset
			if err := cur.Decode(&asset); err != nil {
				continue
			}
			result := models.SearchResult{
				ID:   asset.ID.Hex(),
				Name: asset.Name,
				// Thumbnail:    asset.Thumbnail,
				// ThumbnailExt: asset.ThumbnailExt,
			}
			results = append(results, result)
		}

		if err := cur.Err(); err != nil {
			log.Println("Database query failed :", err)
			c.JSON(http.StatusInternalServerError, responses.AssetResponse{Code: 0, Message: "Error reading from database: " + err.Error()})
			return
		}

		if len(results) == 0 {
			log.Println("Success, but no assets found matching the criteria :", results)
			c.JSON(http.StatusOK, responses.AssetResponse{Code: 1, Message: "Success, but no assets found matching the criteria", Data: results})
		} else {
			log.Println("Success, asset_search :", results)
			c.JSON(http.StatusOK, responses.AssetResponse{Code: 1, Message: "Assets retrieved successfully", Data: results})
		}
	}
}

func DownThumbnail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var results []models.DownThumbnail
		// asset id
		id := c.Query("id")

		filter := bson.M{}
		if id != "" {
			objID, err := primitive.ObjectIDFromHex(id) // 문자열 ID를 ObjectID로 변환
			if err != nil {
				log.Println("Invalid ID format :", err)
				c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "Invalid ID format"})
				return
			}
			filter["_id"] = objID // 정확한 ObjectID로 필터링
		}

		cur, err := assetCollection.Find(ctx, filter)
		if err != nil {
			log.Println("Database query failed :", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "Database query failed"})
			return
		}
		defer cur.Close(ctx)

		for cur.Next(ctx) {
			var asset models.Asset
			if err := cur.Decode(&asset); err != nil {
				continue
			}
			result := models.DownThumbnail{
				ID:           asset.ID.Hex(),
				Thumbnail:    asset.Thumbnail,
				ThumbnailExt: asset.ThumbnailExt,
			}
			results = append(results, result)
		}

		if err := cur.Err(); err != nil {
			log.Println("Error reading from database :", err)
			c.JSON(http.StatusInternalServerError, responses.AssetResponse{Code: 0, Message: "Error reading from database: " + err.Error()})
			return
		}

		if len(results) == 0 {
			log.Println("Success, but no assets found matching :", results)
			c.JSON(http.StatusOK, responses.AssetResponse{Code: 1, Message: "Success, but no assets found matching the criteria", Data: results})
		} else {
			log.Println("Success, DownThumbnail :", results)
			c.JSON(http.StatusOK, responses.AssetResponse{Code: 1, Message: "Assets retrieved successfully", Data: results})
		}
	}
}

func DownAsset() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var results []models.DownFile
		// asset id
		id := c.Query("id")

		filter := bson.M{}
		if id != "" {
			objID, err := primitive.ObjectIDFromHex(id) // 문자열 ID를 ObjectID로 변환
			if err != nil {
				log.Println("Invalid ID format :", err)
				c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "Invalid ID format"})
				return
			}
			filter["_id"] = objID // 정확한 ObjectID로 필터링
		}

		cur, err := assetCollection.Find(ctx, filter)
		if err != nil {
			log.Println("Database query failed :", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "Database query failed"})
			return
		}
		defer cur.Close(ctx)

		for cur.Next(ctx) {
			var asset models.Asset
			if err := cur.Decode(&asset); err != nil {
				continue
			}
			result := models.DownFile{
				ID:   asset.ID.Hex(),
				File: asset.File,
			}
			results = append(results, result)
		}

		if err := cur.Err(); err != nil {
			log.Println("Error reading from database :", err)
			c.JSON(http.StatusInternalServerError, responses.AssetResponse{Code: 0, Message: "Error reading from database: " + err.Error()})
			return
		}

		if len(results) == 0 {
			log.Println("Success, but no assets found matching :", results)
			c.JSON(http.StatusOK, responses.AssetResponse{Code: 1, Message: "Success, but no assets found matching the criteria", Data: results})
		} else {
			log.Println("Success, DownAsset :", results)
			c.JSON(http.StatusOK, responses.AssetResponse{Code: 1, Message: "Assets retrieved successfully", Data: results})
		}
	}
}
