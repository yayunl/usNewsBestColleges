package main

const (
	DefaultSemaphoreSize = 10

	USNewsHTTPGetTimeout    = 10
	WikipediaHTTPGetTimeout = 5

	HTTPReqHeaderUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	USNewsBaseURL = "https://www.usnews.com/best-colleges/api/search?_sort=%s&_sortDirection=asc&_page=%d"
	// SchoolPageCount USNewsBaseURL = "https://www.usnews.com/best-colleges/api/search?_sort=schoolName&_sortDirection=asc&_page=%d"

	DefaultSchoolPageCountStartOffset = 1
	DefaultSchoolPageCountEndOffset   = 183
	TopSchoolsCount                   = 150 // Collect logos for top `TopSchoolsCount` schools

)
