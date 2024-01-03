package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	"usNewsBestColleges/college"
	"usNewsBestColleges/semaphore"
)

func fetchSchools(url string) ([]college.College, error) {
	fmt.Println("Fetching school assets from: ", url)
	client := &http.Client{
		Timeout: USNewsHTTPGetTimeout * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		newErr := fmt.Errorf("failed to create a new request for the url: %s. err: %s", url, err)
		return nil, newErr
	}
	//req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) ")
	req.Header.Set("User-Agent", HTTPReqHeaderUserAgent)
	resp, err := client.Do(req)
	if err != nil {
		newErr := fmt.Errorf("failed to get the response from the url: %s. err: %s", url, err)
		return nil, newErr
	} else {
		//fmt.Println("Successfully fetched school assets from: ", url)
	}
	defer resp.Body.Close()

	var SchoolDataResult college.SchoolData
	err = json.NewDecoder(resp.Body).Decode(&SchoolDataResult)
	if err != nil {
		newErr := fmt.Errorf("failed to decode the response body. the url is %s. err: %s", url, err)
		return nil, newErr
	}
	//fmt.Println("Successfully decoded the response body. ")
	//for _, c := range SchoolDataResult.Data.Items {
	//	fmt.Printf("College name: %+v\n", c.Institution.DisplayName)
	//	fmt.Printf("Tuition: %+v\n", c.SearchData.Tuition.DisplayValue)
	//	fmt.Printf("College detailed page: %s\n", "https://www.usnews.com/best-colleges/"+c.Institution.UrlName+"-"+c.Institution.PrimaryKey)
	//	fmt.Printf("\n\n")
	//}
	return SchoolDataResult.Data.Items, nil
}

func downloadImage(url, destination string) error {
	// Make an HTTP GET request
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download image of url: %s with the destionation of %s. err: %s", url, destination, err)
	}
	defer response.Body.Close()

	// Check if the response status code is OK (200)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download image, status code: %d", response.StatusCode)
	}

	// Create or open the destination file
	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy the contents of the HTTP response body to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	//fmt.Printf("Image downloaded successfully to: %s\n", destination)
	return nil
}

type SchoolLogoResult struct {
	College college.College `json:"college"`
	Err     error           `json:"err"`
}

func fetchLogo(college college.College, url string, logoDestination string, fetchedSchoolLogoChan chan SchoolLogoResult) (any, error) {
	client := &http.Client{
		Timeout: WikipediaHTTPGetTimeout * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		newErr := fmt.Errorf("failed to create a new request for the url: %s. err: %s", url, err)
		fetchedSchoolLogoChan <- SchoolLogoResult{
			College: college,
			Err:     newErr,
		}
		return nil, newErr
	}
	req.Header.Set("User-Agent", HTTPReqHeaderUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		newErr := fmt.Errorf("failed to get the response from the url: %s. err: %s", url, err)
		fetchedSchoolLogoChan <- SchoolLogoResult{
			College: college,
			Err:     newErr,
		}
		return nil, newErr
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		newErr := fmt.Errorf("Error downloading image: %s. url: %s\n", resp.Status, url)
		fetchedSchoolLogoChan <- SchoolLogoResult{
			College: college,
			Err:     newErr,
		}
		return nil, newErr
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fetchedSchoolLogoChan <- SchoolLogoResult{
			College: college,
			Err:     err,
		}
		return nil, err
	}
	//fmt.Printf("doc: %+v\n", doc)
	doc.Find(".infobox.vcard > tbody > tr").Each(func(i int, s *goquery.Selection) {
		imgSrc, exists := s.Find("img").Attr("src")
		if exists {
			imgUrl := "https:" + imgSrc
			err = downloadImage(imgUrl, logoDestination)
			return
		}
		//base64toJpg(imgSrc, schoolName)
	})
	if err != nil {
		newErr := fmt.Errorf("failed to decode the response body. the url is %s. err: %s", url, err)
		fetchedSchoolLogoChan <- SchoolLogoResult{
			College: college,
			Err:     newErr,
		}
		return nil, newErr
	}

	fetchedSchoolLogoChan <- SchoolLogoResult{
		College: college,
		Err:     nil,
	}
	return nil, nil
}

type CollegeResult struct {
	Colleges []college.College `json:"college"`
	Err      error             `json:"err"`
}

func crawl(url string, sem *semaphore.Semaphore) ([]college.College, error) {
	// Fetch the school assets with the semaphore to limit the number of concurrent requests
	schools, err := sem.Process(
		func() (interface{}, error) {
			return fetchSchools(url)
		})
	if err != nil {
		return nil, err
	}
	return schools.([]college.College), nil
}

// collectSchoolData collects school assets using the USNews API. The assets are written to JSON files by rank type.
func collectSchoolData(startOffset, endOffset int, sem *semaphore.Semaphore, saveToJSONFilePath string) {
	// collectSchoolDataWithUSNewsAPI collects school assets using the USNews API. The assets is written to JSON files by rank type.
	// No parameters.
	// No return values.

	// Create the URLs for the school assets
	rankUrls := make([]string, endOffset-startOffset+1)
	for i := startOffset; i <= endOffset; i++ {
		rankUrl := fmt.Sprintf(USNewsBaseURL, "rank", i)
		rankUrls[i-startOffset] = rankUrl
	}
	//fmt.Println("rankUrls: ", rankUrls)

	collegesChan := make(chan CollegeResult)
	// Get the school assets with the USNews API
	var wg0 sync.WaitGroup
	for _, rankUrl := range rankUrls {
		wg0.Add(1)
		if rankUrl == "" {
			continue
		}
		go func(url string) {
			schools, err := crawl(url, sem)
			schoolResult := CollegeResult{
				Colleges: schools,
				Err:      err,
			}
			//fmt.Println("Collected the schoolResult. The first college: ", schoolResult.Colleges[0].Institution.DisplayName)
			collegesChan <- schoolResult
			wg0.Done()

		}(rankUrl)
	}
	go func() {
		wg0.Wait()
		close(collegesChan)
	}()

	// Collect the school assets
	var colleges []college.College
	var errs []error
	for c := range collegesChan {
		//fmt.Printf("Collected school assets. The first college in the list: %+v\n\n", c.Colleges[0].Institution.DisplayName)
		colleges = append(colleges, c.Colleges...)
		errs = append(errs, c.Err)
	}

	// Sort the school assets by rank
	sort.Slice(colleges, func(i, j int) bool {
		return colleges[i].Institution.RankingDisplayRank < colleges[j].Institution.RankingDisplayRank
	})

	var rankTypes []string
	universitiesByRankType := make(map[string][]college.College)

	// Print the school assets
	for i, school := range colleges {
		fmt.Printf("College # %d\n", i+1)
		fmt.Printf("College name: %+v\n", school.Institution.DisplayName)
		fmt.Printf("Tuition: %+v\n", school.SearchData.Tuition.DisplayValue)
		fmt.Printf("Rank %+v in %s\n", school.Institution.RankingDisplayRank, school.Institution.RankingDisplayName)
		fmt.Printf("Enrollment: %+v\n", school.SearchData.Enrollment.DisplayValue)
		fmt.Printf("--------------------\n")
		if !contains(rankTypes, school.Institution.RankingDisplayName) {
			rankTypes = append(rankTypes, school.Institution.RankingDisplayName)
		}
		switch school.Institution.RankingDisplayName {
		case "National Universities":
			universitiesByRankType["nationalUniversities"] = append(universitiesByRankType["nationalUniversities"], school)
		case "National Liberal Arts Colleges":
			universitiesByRankType["nationalLiberalArtsColleges"] = append(universitiesByRankType["nationalLiberalArtsColleges"], school)
		case "Regional Universities South":
			universitiesByRankType["regionalUniversitiesSouth"] = append(universitiesByRankType["regionalUniversitiesSouth"], school)
		case "Regional Universities North":
			universitiesByRankType["regionalUniversitiesNorth"] = append(universitiesByRankType["regionalUniversitiesNorth"], school)
		case "Regional Universities Midwest":
			universitiesByRankType["regionalUniversitiesMidwest"] = append(universitiesByRankType["regionalUniversitiesMidwest"], school)
		case "Regional Universities West":
			universitiesByRankType["regionalUniversitiesWest"] = append(universitiesByRankType["regionalUniversitiesWest"], school)
		case "Regional Colleges South":
			universitiesByRankType["regionalCollegesSouth"] = append(universitiesByRankType["regionalCollegesSouth"], school)
		case "Regional Colleges North":
			universitiesByRankType["regionalCollegesNorth"] = append(universitiesByRankType["regionalCollegesNorth"], school)
		case "Regional Colleges Midwest":
			universitiesByRankType["regionalCollegesMidwest"] = append(universitiesByRankType["regionalCollegesMidwest"], school)
		case "Regional Colleges West":
			universitiesByRankType["regionalCollegesWest"] = append(universitiesByRankType["regionalCollegesWest"], school)
		case "Health Professions Schools":
			universitiesByRankType["healthProfessionsSchools"] = append(universitiesByRankType["healthProfessionsSchools"], school)
		case "Business Schools":
			universitiesByRankType["businessSchools"] = append(universitiesByRankType["businessSchools"], school)
		case "Arts Schools":
			universitiesByRankType["artsSchools"] = append(universitiesByRankType["artsSchools"], school)
		case "Other Schools":
			universitiesByRankType["otherSchools"] = append(universitiesByRankType["otherSchools"], school)
		case "Faith-related Schools":
			universitiesByRankType["faithRelatedSchools"] = append(universitiesByRankType["faithRelatedSchools"], school)
		case "Tribal Schools":
			universitiesByRankType["tribalSchools"] = append(universitiesByRankType["tribalSchools"], school)
		case "Engineering & Technology Schools":
			universitiesByRankType["engineeringAndTechnologySchools"] = append(universitiesByRankType["engineeringAndTechnologySchools"], school)
		case "Medical Schools and Centers":
			universitiesByRankType["medicalSchoolsAndCenters"] = append(universitiesByRankType["medicalSchoolsAndCenters"], school)
		case "Miscellaneous Schools":
			universitiesByRankType["miscellaneousSchools"] = append(universitiesByRankType["miscellaneousSchools"], school)
		case "Research Schools":
			universitiesByRankType["researchSchools"] = append(universitiesByRankType["researchSchools"], school)
		default:
			universitiesByRankType["unknown"] = append(universitiesByRankType["unknown"], school)
		}
	}
	// Write the school assets to JSON files by rank type
	for rankType, universities := range universitiesByRankType {
		_ = writeSortedCollegesByRankToJSON(saveToJSONFilePath+rankType, universities)
	}
	// Print the errors
	for _, err := range errs {
		if err != nil {
			fmt.Printf("Error: %+v\n", err)
		}
	}
	// Print the rank types
	for _, rankType := range rankTypes {
		fmt.Printf("Rank type: %+v\n", rankType)
	}
}

func crawlSchoolLogo(college college.College, url, dest string, sem *semaphore.Semaphore, ch chan SchoolLogoResult) (any, error) {
	// Fetch the school assets with the semaphore to limit the number of concurrent requests
	_, err := sem.Process(func() (interface{}, error) {
		return fetchLogo(college, url, dest, ch)
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func collectSchoolLogs(collegeJsonFile string, topSchoolCountInJsonFile int, sem *semaphore.Semaphore) {
	// Read the school assets from JSON files by rank type
	schools, err := readCollegesFromJSONFile(collegeJsonFile)
	if err != nil {
		fmt.Println("Error reading the school assets from JSON file: ", err)
		return
	}
	// Download the school logos of the top `TopSchoolsCount` universities via Wikipedia
	var wg sync.WaitGroup
	fetchedSchoolLogoChan := make(chan SchoolLogoResult)
	if topSchoolCountInJsonFile < 0 {
		topSchoolCountInJsonFile = len(schools)
	}
	for _, school := range schools[0:topSchoolCountInJsonFile] {

		schoolDisplayName := school.Institution.DisplayName
		if strings.Contains(schoolDisplayName, "-") {
			schoolDisplayName = strings.ReplaceAll(schoolDisplayName, "-", " ")
		}
		schoolPrimaryKey := school.Institution.PrimaryKey

		wikiUrl := "https://en.wikipedia.org/wiki/" + schoolDisplayName
		imgDestination := "./assets/img/" + schoolPrimaryKey + "_" + schoolDisplayName + ".png"

		wg.Add(1)
		go func(college college.College, url, dest string,
			sema *semaphore.Semaphore, fetchedSchoolLogoChan chan SchoolLogoResult) {
			_, err = crawlSchoolLogo(college, url, dest, sema, fetchedSchoolLogoChan)
			if err != nil {
				fmt.Println("Error downloading image: ", err)
			}
			wg.Done()
		}(school, wikiUrl, imgDestination, sem, fetchedSchoolLogoChan)

	}

	go func() {
		wg.Wait()
		close(fetchedSchoolLogoChan)
	}()

	var successSchools []college.College
	var failedSchools []college.College
	for fetchedSchool := range fetchedSchoolLogoChan {
		//fmt.Println("Fetched school img: ", fetchedSchoolLogo)
		if fetchedSchool.Err != nil {
			fmt.Println("Error downloading image: ", fetchedSchool.Err)
			failedSchools = append(failedSchools, fetchedSchool.College)
		} else {
			successSchools = append(successSchools, fetchedSchool.College)
		}
	}
	writeToJSONFile("successSchools", successSchools)
	writeToJSONFile("failedSchools", failedSchools)
}

//func getColleges(colleges []college.College) ([]college.College, error) {
//
//	collegesCpy := make([]college.College, len(colleges))
//	copy(collegesCpy, colleges)
//	for i, c := range colleges {
//		// Iterate over the matches and extract the values
//
//		tuition := c.SearchData.Tuition.DisplayValue.(string)
//
//		if strings.Contains(tuition, "state") {
//			var tuitionStr string
//			// Define the regular expression pattern to match the values
//			pattern := `(?:name:\()?\s*([a-zA-Z-]+)\s*(?:\))?\s*value:\s*\$([0-9,]+)`
//			// Compile the regular expression pattern
//			re := regexp.MustCompile(pattern)
//			// Find all matches in the input string
//			matches := re.FindAllStringSubmatch(tuition, -1)
//			for _, match := range matches {
//				if len(match) == 3 {
//					key := match[1]
//					value := match[2]
//					tuitionStr += fmt.Sprintf("%s: $%s\n", key, value)
//				}
//			}
//			tuition = tuitionStr
//		}
//		collegesCpy[i].SearchData.Tuition.DisplayValue = tuition
//	}
//	return collegesCpy, nil
//}

func getCollegeLogos(colleges []college.College) (map[string]string, error) {
	logos := make(map[string]string)
	// Get logs from the assets/img folder
	files, err := ioutil.ReadDir("./assets/img")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		fileName := file.Name()
		fileNameSplit := strings.Split(fileName, "_")
		logos[fileNameSplit[0]] = fmt.Sprintf("./assets/img/%s", fileName)
	}
	return logos, nil
}

func main() {
	//// Part 1: Collect school data and logs
	//// Create a semaphore with a buffer
	//sem := semaphore.NewSemaphore(DefaultSemaphoreSize)
	//// Collect school data from the USNews API and save the data to JSON files by the rank of the school
	//startOffset, endOffset := DefaultSchoolPageCountStartOffset, DefaultSchoolPageCountEndOffset
	//collectSchoolData(startOffset, endOffset, sem, "./assets/data/")
	//// Collect school logos from Wikipedia
	//collectSchoolLogs("./failedSchools.json", -1, sem)

	// Part 2: Run the web server to display the school data
	// Create a new Gin router
	router := gin.Default()

	// Load the HTML templates
	router.SetFuncMap(template.FuncMap{
		"capitalize": Capitalize,
		"split":      SplitStringByDelimiter,
	})
	router.LoadHTMLGlob("templates/*")

	// Load the static assets
	router.Static("/assets", "./assets")
	router.Static("/static", "./static/")

	// Define routes
	router.GET("/", nationalUniversitiesHandler)

	// Start the server
	router.Run(":8000")

}

func nationalUniversitiesHandler(c *gin.Context) {
	var colleges []college.College
	var err error
	// Read the school data from JSON files
	colleges, err = readCollegesFromJSONFile("./assets/data/nationalUniversities.json")
	//colleges, err = getColleges(schools)
	// Get the logo of the college
	logos, err := getCollegeLogos(colleges)

	if err != nil {
		fmt.Println("Error getting colleges: ", err)
		return
	}

	c.HTML(200, "index.html", gin.H{
		"title":    "Best National Universities",
		"colleges": colleges,
		"logos":    logos,
	})

}
