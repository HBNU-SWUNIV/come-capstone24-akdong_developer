package controllers

import (
	"context"
	"fmt"
	"log"
	"map_http_go/responses"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// map_id값으로 time반환
func Get_time() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		mapId, _ := strconv.ParseFloat(c.Query("mapId"), 64)
		fmt.Println(mapId)

		filter := bson.M{
			"map_id":   mapId,
			"chunkNum": 0,
		}

		projection := bson.M{
			"_id":      0,
			"mapCTime": 1,
		}

		cursor, err := mapCollection.Find(ctx, filter, options.Find().SetProjection(projection))
		if err != nil {
			
			c.JSON(http.StatusInternalServerError, responses.MapResponse{Code: 0, Message: "error"})
			log.Println("(time)FindErr :", err)
			return
		}

		fmt.Println(cursor)

		var results []bson.M
		if err = cursor.All(ctx, &results); err != nil {
			
			c.JSON(http.StatusInternalServerError, responses.MapResponse{Code: 0, Message: "error"})
			log.Println("(time)ReturnErr :", err)
			return
		}

		// 결과가 없으면 에러 반환
		if len(results) == 0 {
			
			c.JSON(http.StatusNotFound, responses.MapResponse{Code: 0, Message: "No data found"})
			log.Println("(time)NoDataErr :", err)
			return
		}

		fmt.Println(results)
		str_results := results[0]["mapCTime"].(string)

		// JSON으로 변환된 시간 반환
		
		c.JSON(http.StatusOK, responses.MapResponse{Code: 1, Message: str_results})
		log.Println("(time)Success :", mapId)

	}

}
