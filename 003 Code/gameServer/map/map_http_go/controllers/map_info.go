package controllers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"map_http_go/responses"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GET은 반대로 map안에 데이터를 저장하고 json으로 보내주면 되지 않을까?
// query는 어떻게 전송해야하나(router에서는 /만처리해주고 Query를 통해서 매개변수 확인)
// 자꾸 collection 찾을 수 없다고 뜸(형 변환 문제)
func GetMap() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// // url에서 매개변수 추출
		mapId, _ := strconv.ParseFloat(c.Query("mapID"), 64) // double로 저장되어 있는데 .Atoi(int변환)으로 해서 반환 못하는 문제 발생 (해결완)
		mapVersion, _ := strconv.ParseFloat(c.Query("version"), 64)
		mapChunk, _ := strconv.ParseFloat(c.Query("chunk"), 64)

		// MongoDB에서 데이터 조회
		mapinfo := make(map[string]interface{})

		filter := bson.M{
			"map_id":   mapId,
			"version":  mapVersion,
			"chunkNum": mapChunk,
		}

		// fmt.Println("filter", filter)

		// ObjectId 제거
		err := mapCollection.FindOne(ctx, filter).Decode(&mapinfo)
		// fmt.Println("mapinfo", mapinfo)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				
				c.JSON(http.StatusNotFound, responses.MapResponse{Code: 0, Message: "No documents found"})
				log.Println("(info)NoDocumentsErr :",err)
				return
			}
			c.JSON(http.StatusInternalServerError, responses.MapResponse{Code: 0, Message: err.Error()})
			return
		}

		// 특정 key 제외
		filtermapinfo := make(map[string]interface{})

		for key, value := range mapinfo {
			if key != "_id" {
				filtermapinfo[key] = value
			}
		}

		
		// 데이터 반환
		c.JSON(http.StatusOK, responses.MapResponse_map{Code: 1, Message: filtermapinfo}) // 형변환
		log.Println("(info)InfoSuccess :",mapId)
	}
}
