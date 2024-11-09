package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"map_http_go/responses"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetList() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 표시되는 필드만 표현
		projection := bson.M{
			"_id":     0,
			"map_id":  1,
			"mapName": 1,
		}

		// filter 모두 존재하는 경우에만 출력
		filter := bson.M{
			"$and": []bson.M{
				{"map_id": bson.M{"$exists": true}},
				{"mapName": bson.M{"$exists": true}},
				{"chunkNum": 0},
			},
		}

		// collection.Find(context:취소 시그널 및 타임아웃 전달, 빈맵 : 모든 문서 선택, setProjection(검색 옵션 설정))
		// cursor : 결과 집합의 다음 항목 가져올 수 있음
		//cursor, err := mapCollection.Find(ctx, bson.M{}, options.Find().SetProjection(projection)) // rejection이랑 filter 정확하게 알기
		cursor, err := mapCollection.Find(ctx, filter, options.Find().SetProjection(projection))
		if err != nil {
			
			c.JSON(http.StatusInternalServerError, responses.MapResponse{Code: 0, Message: "error"})
			log.Println("(list)FindErr :", err)
			return
		}

		fmt.Println(cursor)

		// cursor에서 반환된 모든 값을 가져와 map[string]interface{} 슬라이스로 변환
		var results []map[string]interface{} // 여러개라서 []map[string]interface{}
		if err = cursor.All(ctx, &results); err != nil {
			
			c.JSON(http.StatusInternalServerError, responses.MapResponse{Code: 0, Message: "error"})
			log.Println("(list)FindReturnErr :", err)
			return
		}

		// JSON으로 결과 반환
		
		c.JSON(http.StatusOK, responses.MapResponse_list{Code: 1, Message: results})
		log.Println("(list)Success")
	}
}
