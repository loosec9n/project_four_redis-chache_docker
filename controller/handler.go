package controller

import (
	"Microservices_Go_Caching_with_Redis/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

func (a *Api) RedisHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("In the handler")

	//something to get the data back
	//can be sql  or 3rd part API's

	q := r.URL.Query().Get("query")

	data, cacheHit, err := a.GetURLData(r.Context(), q)
	if err != nil {
		log.Println("error no data from getData", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "error no data from getData",
			"result":  err,
		})
		return
	}

	resp := models.APIResponse{
		Cache: cacheHit,
		Data:  data,
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Println("error encoding the respose:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "error encoding the response",
			"result":  err,
		})
		return
	}
}
func (a *Api) GetURLData(ctx context.Context, q string) ([]models.NominatimResponse, bool, error) {

	//is querry cashed

	value, err := a.Cache.Get(ctx, q).Result()
	if err == redis.Nil {
		// want to set call external data source
		escapedQ := url.PathEscape(q)

		address := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json", escapedQ)

		resp, err := http.Get(address)

		if err != nil {
			return nil, false, err
		}

		data := make([]models.NominatimResponse, 0)

		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, false, err
		}

		b, err := json.Marshal(data)
		if err != nil {
			return nil, false, err
		}

		//set the value
		err = a.Cache.Set(ctx, q, bytes.NewBuffer(b).Bytes(), time.Second*15).Err() //15 seconds for the testinf the cashing happens or not. relaistically this will be much higher
		if err != nil {
			return nil, false, err
		}

		//return the response
		return data, false, nil

	} else if err != nil {
		log.Println("error calling redis", err)
		return nil, false, err
	} else {
		//cache hit
		data := make([]models.NominatimResponse, 0)
		//build response

		err := json.Unmarshal([]byte(value), &data)
		if err != nil {
			return nil, false, err
		}

		//return response
		return data, true, nil
	}

}

type Api struct {
	Cache *redis.Client
}

func NewAPI() *Api {

	redisAddress := fmt.Sprintf("%s:6379", os.Getenv("REDIS_URL"))

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return &Api{
		Cache: rdb,
	}
}
