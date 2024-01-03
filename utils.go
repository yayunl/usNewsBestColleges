package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang.org/x/image/draw"
	"image"
	"image/color"

	//"image/draw"
	"image/png"
	"os"
	"sort"
	"strconv"
	"strings"
	"usNewsBestColleges/college"
)

func contains(types []string, name string) bool {
	for _, t := range types {
		if t == name {
			return true
		}
	}
	return false
}

func writeSortedCollegesByRankToJSON(filename string, data any) error {
	colleges := data.([]college.College)
	var college1Rank, college2Rank int
	// Sort the school assets by rank
	sort.Slice(colleges, func(i, j int) bool {
		if college1RankContainsSharp := strings.Contains(colleges[i].Institution.RankingDisplayRank, "#"); !college1RankContainsSharp {
			return false
		}
		if college2RankContainsSharp := strings.Contains(colleges[j].Institution.RankingDisplayRank, "#"); !college2RankContainsSharp {
			return false
		}

		if strings.Contains(colleges[i].Institution.RankingDisplayRank, "-") {
			college1RankStr := strings.Split(strings.Split(colleges[i].Institution.RankingDisplayRank, "-")[0], "#")[1]
			college1Rank, _ = strconv.Atoi(college1RankStr)
		} else {
			college1Rank, _ = strconv.Atoi(strings.Split(colleges[i].Institution.RankingDisplayRank, "#")[1])
		}

		if strings.Contains(colleges[j].Institution.RankingDisplayRank, "-") {
			college2RankStr := strings.Split(strings.Split(colleges[j].Institution.RankingDisplayRank, "-")[0], "#")[1]
			college2Rank, _ = strconv.Atoi(college2RankStr)
		} else {
			college2Rank, _ = strconv.Atoi(strings.Split(colleges[j].Institution.RankingDisplayRank, "#")[1])
		}

		return college1Rank < college2Rank
	})
	// Write the school assets to a JSON file
	file, err := os.Create(filename + ".json")
	if err != nil {
		return err
	}
	defer file.Close()

	// Overwrite the existing file
	err = os.Truncate(file.Name(), 0)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(colleges)
	if err != nil {
		return err
	}
	return nil
}

func writeToJSONFile(filename string, data any) {
	file, err := os.Create(filename + ".json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Overwrite the existing file
	err = os.Truncate(file.Name(), 0)
	if err != nil {
		panic(err)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(data)
	if err != nil {
		panic(err)
	}
}

func readCollegesFromJSONFile(filename string) ([]college.College, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var colleges []college.College
	err = json.NewDecoder(file).Decode(&colleges)
	if err != nil {
		return nil, err
	}
	return colleges, nil
}

// Given a base64 string of a JPEG, encodes it into an JPEG image test.jpg
func base64toJpg(inputData string, outputFileName string) {

	// Decode base64 string
	base64ImgData := strings.Split(inputData, ",")[1]
	data, err := base64.StdEncoding.DecodeString(base64ImgData)
	if err != nil {
		fmt.Println("Error decoding base64:", err)
		return
	}

	// Decode PNG assets
	img, _, err := image.Decode(strings.NewReader(string(data)))
	if err != nil {
		fmt.Println("Error decoding PNG:", err)
		return
	}

	// Create a new JPG file
	jpgFile, err := os.Create(outputFileName + ".jpg")
	if err != nil {
		fmt.Println("Error creating JPG file:", err)
		return
	}
	defer jpgFile.Close()

	// Encode and save as JPG
	err = png.Encode(jpgFile, img)
	if err != nil {
		fmt.Println("Error encoding and saving as JPG:", err)
		return
	}

	fmt.Println("Image saved as output.jpg")

}

func pngToBase64(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Println("Error decoding image:", err)
		return ""
	}

	// Create a new image with the desired resolution and white background
	newImg := image.NewRGBA(image.Rect(0, 0, 120, 120))
	draw.Draw(newImg, newImg.Bounds(), &image.Uniform{color.White}, image.ZP, draw.Src)

	// Resize the image to the new resolution
	draw.ApproxBiLinear.Scale(newImg, newImg.Bounds(), img, img.Bounds(), draw.Over, nil)

	// Encode the image to PNG
	buf := new(bytes.Buffer)
	err = png.Encode(buf, newImg)
	if err != nil {
		fmt.Println("Error encoding image:", err)
		return ""
	}

	// Create a base64 string of the image
	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	return base64Str
}

func Capitalize(s string) string {
	s = strings.ToLower(s)
	return strings.ToUpper(s[:1]) + s[1:]
}

func SplitStringByDelimiter(s string) []string {
	return strings.Split(s, "%")
}
