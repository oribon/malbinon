package dogsapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func GenerateDogName() (string, error) {
	rand.Seed(time.Now().UTC().UnixNano())
	var dogBreeds []string
	response, err := http.Get("https://dog.ceo/api/breeds/list/all")
	if err != nil {
		return "", err
	} else {
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return "", err
		}
		defer response.Body.Close()
		var objmap map[string]map[string][]string
		_ = json.Unmarshal(data, &objmap)
		for breed, subBreeds := range objmap["message"] {
			if len(subBreeds) == 0 {
				dogBreeds = append(dogBreeds, breed)
			} else {
				for _, subBreed := range subBreeds {
					dogBreeds = append(dogBreeds, breed+"-"+subBreed)
				}
			}
		}
	}

	return dogBreeds[rand.Intn(len(dogBreeds))] + strconv.Itoa(rand.Intn(10)) + strconv.Itoa(rand.Intn(10)), nil
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	var dogBreeds []string
	response, err := http.Get("https://dog.ceo/api/breeds/list/all")
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println(err)
		}
		defer response.Body.Close()
		var objmap map[string]map[string][]string
		_ = json.Unmarshal(data, &objmap)
		for breed, subBreeds := range objmap["message"] {
			if len(subBreeds) == 0 {
				dogBreeds = append(dogBreeds, breed)
			} else {
				for _, subBreed := range subBreeds {
					dogBreeds = append(dogBreeds, breed+"-"+subBreed)
				}
			}
		}
		fmt.Println(dogBreeds[rand.Intn(len(dogBreeds))] + strconv.Itoa(rand.Intn(10)) + strconv.Itoa(rand.Intn(10)))
	}
}
