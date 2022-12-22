package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"io/ioutil"
	//"io"
	"log"
	"net/http"
	//"os"
	_ "image/jpeg"
	"strconv"
	"strings"
)

type Services []Service
type Service struct {
	Name           string         // example: "GitHub"
	UserExistsFunc UserExistsFunc // example: SimpleUserExistsCheck()
	GetInfoFunc    GetInfoFunc    // example: EmptyAccountInfo()
	BaseUrl        string         // example: "https://github.com"
}

type Accounts map[string]Account
type Account struct {
	Service  string   `json:"service"`  // example: GitHub
	Id       string   `json:"id"`       // example: 1224234
	Username string   `json:"username"` // example: 9glenda
	Url      string   `json:"url"`      // example: https://github.com/9glenda
	Picture  []string `json:"profilePicture"`
	ImgHash  []uint64 `json:"imgHash"`
	Bio      []string `json:"bio"` // example: pro hacka
}

type GetInfoFunc func(string, Service) Account // (username)
type UserExistsFunc func(string, string) bool  // (BaseUrl,username)

var DefaultServices = Services{
	Service{
		Name:           "GitHub",
		UserExistsFunc: SimpleUserExistsCheck,
		GetInfoFunc:    GithubInfo,
		BaseUrl:        "https://github.com/",
	},
	Service{
		Name:           "SlideShare",
		UserExistsFunc: SimpleUserExistsCheck,
		GetInfoFunc:    SlideshareInfo,
		BaseUrl:        "https://slideshare.net/",
	},
}

func SimpleUserExistsCheck(BaseUrl, username string) bool {
	return GetStatusCode(BaseUrl+username) == 200
}

func EmptyAccountInfo(username string, service Service) Account {
	return Account{
		Service:  service.Name,
		Username: username,
		Bio:      nil,
	}
}

// maybe remove
func DefaultServicesHandler(username string) Accounts {
	return ServicesHandler(DefaultServices, username)
}

func HttpRequest(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	return string(body)
}

func ServicesHandler(servicesToCheck Services, username string) Accounts {
	var accounts = make(Accounts)
	for i := 0; i < len(servicesToCheck); i++ {
		service := servicesToCheck[i]
		if service.UserExistsFunc(service.BaseUrl, username) {
			accounts[service.Name] = service.GetInfoFunc(username, service)
		}
	}
	return accounts
}
func getImg(img string) image.Image {
	reader := strings.NewReader(img)
	decodedImg, imgType, err := image.Decode(reader)
	log.Printf("image type:%s", imgType)
	if err != nil {
		log.Println(err)
	}
	return decodedImg
}

func EncodeBase64(img string) string {
	decodedImg := getImg(img)
	buf := new(bytes.Buffer)
	err := png.Encode(buf, decodedImg)
	if err != nil {
		log.Println(err)
	}

	base64Img := base64.StdEncoding.EncodeToString(buf.Bytes())
	return base64Img
}

func SlideshareInfo(username string, service Service) Account {
	avatar_url := "https://cdn.slidesharecdn.com/profile-photo-" + username + "-96x96.jpg"
	log.Printf("avatar_url: %s", avatar_url)

	account := Account{
		Service:  service.Name,
		Username: username,
		Url:      service.BaseUrl + username,
		//Picture: []string{EncodeBase64("https://www.tutorialspoint.com/html/images/test.png")},
	}
	if GetStatusCode(avatar_url) == 200 {
		avatar := HttpRequest(avatar_url)
		account.Picture = []string{EncodeBase64(avatar)} // img := HttpRequest(url)
		account.ImgHash = []uint64{MkImgHash(getImg(avatar))}
	}
	return account
}
func GithubInfo(username string, service Service) Account {
	var data struct {
		Id         int    `json:"id"`
		Bio        string `json:"bio"`
		Avatar_url string `json:"avatar_url"`
		Url        string `json:"html_url"`
	}
	jsonData := HttpRequest("https://api.github.com/users/" + username)
	err := json.Unmarshal([]byte(jsonData), &data)
	log.Printf("avatar_url: %s", data.Avatar_url)
	if err != nil {
		log.Println(err)
	}
	avatar := HttpRequest(data.Avatar_url)
	account := Account{
		Service:  service.Name,
		Username: username,
		Url:      data.Url,
		Id:       strconv.Itoa(data.Id),
		Bio:      []string{data.Bio},
		Picture:  []string{EncodeBase64(avatar)},
		ImgHash:  []uint64{MkImgHash(getImg(avatar))},
	}
	return account
}