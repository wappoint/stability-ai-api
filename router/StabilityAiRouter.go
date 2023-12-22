package router

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
)

var apiKey = "sk-Z0"

const AuthHeaderPrefix = "Bearer "
const ApiHost = "https://api.stability.ai"

type TextToImageImage struct {
	Base64       string `json:"base64"`
	Seed         uint32 `json:"seed"`
	FinishReason string `json:"finishReason"`
}
type TextToImageResponse struct {
	Images []TextToImageImage `json:"artifacts"`
}
type ImageToImageImage struct {
	Base64       string `json:"base64"`
	Seed         uint32 `json:"seed"`
	FinishReason string `json:"finishReason"`
}

type ImageToImageResponse struct {
	Images []ImageToImageImage `json:"artifacts"`
}

type ImageToVideoResponse struct {
	Video        string `json:"base64"`
	Seed         uint32 `json:"seed"`
	FinishReason string `json:"finishReason"`
}

func StabilityAiRouter(router *fiber.App) {
	//v1Text2Image()
	router.Get("/userInfo", func(c *fiber.Ctx) error {
		body := GetUserInfo()
		return c.SendString(body)
	})
	router.Get("/engines", func(c *fiber.Ctx) error {
		body := getEngines()
		return c.SendString(body)
	})
	router.Get("/balance", func(c *fiber.Ctx) error {
		body := getBalance()
		return c.SendString(body)
	})
	router.Get("/text2image", func(c *fiber.Ctx) error {
		text := ""
		body := v1Text2Image(text)
		return c.Render("template/text2image.html", &body)
	})
	router.Get("/image2image", func(c *fiber.Ctx) error {
		inputImagePath := "D:\\input_v1_image2img_%d.png"
		outPutImagePath := "D:\\output_v1_image2img_%d.png"
		body := v1Image2Image(inputImagePath, outPutImagePath)
		return c.Render("template/image2image.html", &body)
	})
	router.Get("/image2video", func(c *fiber.Ctx) error {
		inputImagePath := "D:\\input_v2_image2video.png"
		body := v2Image2Video(inputImagePath)
		return c.SendString(body)
	})
	router.Get("/image2videobyid", func(c *fiber.Ctx) error {
		videoId := "68a8b06c5601d4ad07475652392d41aafe14913d886fd169f5b8f9b82ca10118"
		outPutVideoPath := "D:\\output_v2_image2video.mp4"
		v2Image2VideoById(videoId, outPutVideoPath)
		return c.SendString("succeed video , please check path -> " + outPutVideoPath)
	})
}

func GetUserInfo() string {
	url := "/v1/user/account"
	reqUrl := ApiHost + url
	body, res := GetRequestDefault(reqUrl)
	if res.StatusCode == 200 {
		println(string(body))
	} else {
		panic("Non-200 response: " + string(body))
	}
	return string(body)
}

func getEngines() string {
	url := "/v1/engines/list"
	reqUrl := ApiHost + url
	body, res := GetRequestDefault(reqUrl)
	if res.StatusCode == 200 {
		println("engines -> " + string(body))
	} else {
		panic("Non-200 response: " + string(body))
	}
	return string(body)
}

func getBalance() string {
	url := "/v1/user/balance"
	reqUrl := ApiHost + url
	body, res := GetRequestDefault(reqUrl)
	if res.StatusCode == 200 {
		println("getBalance -> " + string(body))
	} else {
		panic("Non-200 response: " + string(body))
	}
	return string(body)
}

func v1Text2Image(text string) map[int]string {
	engineId := "stable-diffusion-512-v2-1"
	url := "/v1/generation/" + engineId + "/text-to-image"
	var data = []byte(`{
		"text_prompts": [
		  {
			"text": ` + text + `
		  }
		],
		"cfg_scale": 7,
		"height": 1024,
		"width": 1024,
		"samples": 1,
		"steps": 30
  	}`)
	images := v1Text2Images(url, data)
	return images
}
func v1Text2Images(url string, data []byte) map[int]string {
	content, res := PostRequestInfo(url, data)
	if res.StatusCode != 200 {
		var body map[string]interface{}
		if err := json.NewDecoder(bytes.NewReader(content)).Decode(&body); err != nil {
			panic(err)
		}
		panic(fmt.Sprintf("Non-200 response: %s", body))
	}
	// Decode the JSON body
	var body TextToImageResponse
	if err := json.NewDecoder(bytes.NewReader(content)).Decode(&body); err != nil {
		panic(err)
	}
	imageLen := len(body.Images)
	fmt.Println("images len -> " + strconv.Itoa(imageLen))
	// write file
	newFilePath := "D:\\v1_txt2img_%d.png"
	imagesBytes := writeText2ImageFiles(body, newFilePath)
	return imagesBytes
}

func v1Image2Image(imagePath string, newImagePath string) map[int]string {
	engineId := "stable-diffusion-xl-1024-v1-0"
	url := "/v1/generation/" + engineId + "/image-to-image"
	reqUrl := ApiHost + url
	data := &bytes.Buffer{}
	writer := multipart.NewWriter(data)

	// Write the init image to the request
	initImageWriter, _ := writer.CreateFormField("init_image")
	initImageFile, initImageErr := os.Open(imagePath)
	if initImageErr != nil {
		panic("Could not open init_image.png")
	}
	_, _ = io.Copy(initImageWriter, initImageFile)

	// Write the options to the request
	_ = writer.WriteField("init_image_mode", "IMAGE_STRENGTH")
	_ = writer.WriteField("image_strength", "0.35")
	_ = writer.WriteField("text_prompts[0][text]", "Galactic dog with a cape")
	_ = writer.WriteField("cfg_scale", "7")
	_ = writer.WriteField("samples", "1")
	_ = writer.WriteField("steps", "30")
	writer.Close()
	headers := map[string]string{
		"Content-Type":  writer.FormDataContentType(),
		"Accept":        "application/json",
		"Authorization": AuthHeaderPrefix + apiKey,
	}
	content, res := PostRequest(reqUrl, data.Bytes(), headers)
	if res.StatusCode != 200 {
		var body map[string]interface{}
		if err := json.NewDecoder(bytes.NewReader(content)).Decode(&body); err != nil {
			panic(err)
		}
		panic(fmt.Sprintf("Non-200 response: %s", body))
	}
	// Decode the JSON body
	var body ImageToImageResponse
	if err := json.NewDecoder(bytes.NewReader(content)).Decode(&body); err != nil {
		panic(err)
	}
	imageLen := len(body.Images)
	fmt.Println("images len -> " + strconv.Itoa(imageLen))

	newFilePath := "D:\\v1_image2img_%d.png"
	imagesBytes := writeImage2ImageFiles(body, newFilePath)
	return imagesBytes
}

func v2Image2Video(imagePath string) string {
	url := "/v2alpha/generation/image-to-video"
	reqUrl := ApiHost + url
	file, err := os.Open(imagePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	// fileContent := readFile(imagePath)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	//writer.WriteField("seed", "0")
	//writer.WriteField("cfg_scale", "2.5")
	//writer.WriteField("motion_bucket_id", "40")
	//writer.WriteField("Content-Type", "image/png")
	//part, err := writer.CreateFormFile("image", filepath.Base(file.Name()))
	//if err != nil {
	//	panic(err)
	//}
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			"image", filepath.Base(file.Name())))
	h.Set("Content-Type", "image/png")
	part, err := writer.CreatePart(h)
	_, err = io.Copy(part, file)
	err = writer.Close()
	if err != nil {
		panic(err)
	}
	headers := map[string]string{
		"Content-Type":  writer.FormDataContentType(),
		"Authorization": AuthHeaderPrefix + apiKey,
	}
	content, _ := PostRequest(reqUrl, body.Bytes(), headers)
	println("video -> " + string(content))
	return string(content)
}

//func appendPayload() {
//	boundary := "ASSDFWDFBFWEFWWDF"
//	// HTTP 拼装开始
//	picData := "--" + boundary + "\r\n"
//	// HTTP 文本组装
//	picData = picData + "Content-Disposition:form-data;name=\"meta\"\r\nContent-Type:application/json\r\n\r\n"
//	picData = picData + string(bodyBytes) + "\r\n"
//	picData = picData + "--" + boundary + "\r\n"
//	// HTTP 文件组装
//	picData = picData + "Content-Disposition: form-data; name=\"file\"; filename=\"" + picFilePath + "\"\r\n" + "Content-Type: " + fileContentType + "\r\n\r\n"
//	picData = picData + string(picBytes) + "\r\n"
//	//最后的boundary 尾部会加2个-
//	picData = picData + "--" + boundary + "--"
//}

func v2Image2VideoById(videoId string, outPutFilePath string) {
	url := "/v2alpha/generation/image-to-video/result/" + videoId
	reqUrl := ApiHost + url
	videoBytes, res := GetRequestDefault(reqUrl)
	if res.StatusCode == 200 {
		writeFile(outPutFilePath, videoBytes)
	}
}

func writeText2ImageFiles(body TextToImageResponse, newFilePath string) map[int]string {
	imagesBytes := make(map[int]string)
	// Write the images to disk
	for i, image := range body.Images {
		filePath := fmt.Sprintf(newFilePath, i)
		//fmt.Println("image -> " + image.Base64)
		imageBytes, err := base64.StdEncoding.DecodeString(image.Base64)
		if err != nil {
			panic(err)
		}
		writeFile(filePath, imageBytes)
		imagesBytes[i] = "data:image/png;base64," + image.Base64
	}
	//fmt.Printf("imagesBytes -> %v", imagesBytes)
	return imagesBytes
}
func writeImage2ImageFiles(body ImageToImageResponse, newFilePath string) map[int]string {
	imagesBytes := make(map[int]string)
	// Write the images to disk
	for i, image := range body.Images {
		filePath := fmt.Sprintf(newFilePath, i)
		//fmt.Println("image -> " + image.Base64)
		imageBytes, err := base64.StdEncoding.DecodeString(image.Base64)
		if err != nil {
			panic(err)
		}
		writeFile(filePath, imageBytes)
		imagesBytes[i] = "data:image/png;base64," + image.Base64
	}
	//fmt.Printf("imagesBytes -> %v", imagesBytes)
	return imagesBytes
}

func writeFile(filePath string, content []byte) {
	outFile := fmt.Sprintf(filePath)
	file, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	if _, err := file.Write(content); err != nil {
		panic(err)
	}
	if err := file.Close(); err != nil {
		panic(err)
	}
}

func readFile(filePath string) []byte {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	return content
}

func GetRequestDefault(reqUrl string) (content []byte, res *http.Response) {
	return GetRequestMethod("GET", reqUrl)
}
func GetRequestMethod(method string, reqUrl string) (content []byte, res *http.Response) {
	return RequestData(method, reqUrl, nil)
}

func RequestData(method string, reqUrl string, body []byte) (content []byte, res *http.Response) {
	headers := make(map[string]string)
	headers["Authorization"] = AuthHeaderPrefix + apiKey
	return RequestHeader(method, reqUrl, body, headers)
}

func Request(method string, reqUrl string, body []byte) (content []byte, res *http.Response) {
	return RequestHeader(method, reqUrl, body, nil)
}

func RequestHeader(method string, reqUrl string, body []byte, headers map[string]string) (content []byte, res *http.Response) {
	request, _ := http.NewRequest(method, reqUrl, bytes.NewBuffer(body))
	for k, v := range headers {
		request.Header.Add(k, v)
	}
	res, _ = http.DefaultClient.Do(request)
	defer res.Body.Close()
	content, _ = io.ReadAll(res.Body)
	return content, res
}

func PostRequest(url string, data []byte, headers map[string]string) (content []byte, res *http.Response) {
	return RequestHeader("POST", url, data, headers)
}

func PostRequestInfo(url string, data []byte) (content []byte, res *http.Response) {
	reqUrl := ApiHost + url
	headers := make(map[string]string)
	headers["Authorization"] = AuthHeaderPrefix + apiKey
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/json"
	return PostRequest(reqUrl, data, headers)
}
