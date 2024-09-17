package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"map_http_go/configs"
	"map_http_go/responses"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var mapCollection *mongo.Collection = configs.GetCollection(configs.DB, "map")

// mapCTime 전송하기 (timestamp)

// json 파일 DB 저장
func CreateMap() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 동적으로 Key-Value 쌍 저장 가능
		mapdata := make(map[string]interface{}) // 빈 맵 생성(동적으로 key-value쌍 저장 가능)
		// var map_data models.MapData // map 함수 이름 존재하므로 map쓰면 안됨 + 동적 저장이므로 struct 필요 없음

		//body 유효성 검증
		if err := c.BindJSON(&mapdata); err != nil {
			c.JSON(http.StatusBadRequest, responses.MapResponse{Code: 0, Message: "body error"})
			log.Println("(create)BindJsonErr :", err)
			return
		}

		// fmt.Println("2222222222222", mapdata) // 어떤거 나오는지 비교, json 받아온 값 확인

		// mapdata에서 map_id, version, chunkNum 확인
		mapId := mapdata["map_id"].(float64)
		mapVersion := mapdata["version"].(float64)
		mapChunk := mapdata["chunkNum"].(float64)

		// fmt.Println("3333333333333", mapId, mapVersion, mapChunk)

		// 중복 확인을 위한 filter
		filter := bson.M{
			"$and": []bson.M{ // &인지 $인지 잘 구분할것
				{"map_id": mapId},
				{"version": mapVersion},
				// {"map_id": mapId},
				// {"version": mapVersion},
				{"chunkNum": mapChunk},
			},
		}

		// fmt.Println(filter)

		var existingData map[string]interface{}
		err := mapCollection.FindOne(ctx, filter).Decode(&existingData) // 다른 map으로 만들것, 계속 같은 mapdata(들어온 값)사용해서 오류남, existingData(확인하는 값)

		// 중복 없을 때 Insert
		if err != nil {
			if err == mongo.ErrNoDocuments {
				_, err := mapCollection.InsertOne(ctx, mapdata) // DB에 바로 저장
				if err != nil {
					
					c.JSON(http.StatusInternalServerError, responses.MapResponse{Code: 0, Message: "insert error"})
					log.Println("(create)InsertErr :", err)
					return
				}
				
				c.JSON(http.StatusCreated, responses.MapResponse{Code: 1, Message: "insert success"})
				log.Println("Insert Success :", mapId)
				return
			} else {
				
				c.JSON(http.StatusInternalServerError, responses.MapResponse{Code: 0, Message: err.Error()})
				log.Println("(create)ExistInsertErr :", err)
				return
			}
		}

		// Update(모두 삭제 하고 Insert)
		// 모두 삭제(if chunkNum이 더 있으면)
		_, err2 := mapCollection.ReplaceOne(ctx, filter, mapdata)
		if err2 != nil {
			
			c.JSON(http.StatusInternalServerError, responses.MapResponse{Code: 0, Message: "error"})
			log.Println("(create)UpdateErr :", err)
			return
		}

		// 만약 map_id, version, chunk(개수), mapdata 제외하고 map_id, version있는거 모두 삭제
		// 만약 짧은 정보가 Update 된다면 나머지 Chunk_Num을 삭제
		// 첫번째가 중복되면 이후 chunk_num 모두 삭제하기, 그러면 다음부터는 create가능,
		// -> chunk_num이 들어온 값보다 큰 값이 있으면 삭제
		filterDelete := bson.M{
			"$and": []bson.M{
				{"map_id": mapId},
				{"version": mapVersion},
				{"chunkNum": bson.M{"$gt": mapChunk}},
			},
		}

		_, chk_err := mapCollection.DeleteMany(ctx, filterDelete)
		if chk_err != nil {
			
			c.JSON(http.StatusInternalServerError, responses.MapResponse{Code: 0, Message: "delete err"})
			log.Println("(create)DeleteErr :", chk_err)
		}

		
		c.JSON(http.StatusOK, responses.MapResponse{Code: 1, Message: "update success"})
		log.Println("(create)Update Success :", mapId)

	}
}
