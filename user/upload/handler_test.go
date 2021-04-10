package upload

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/maedu/mtg-cards/server"
)

func testSetup(r *gin.Engine) {
	r.POST("/parse-test", handleParseForTest)
}

func handleParseForTest(c *gin.Context) {
	cards, err := parseUserCards(c)
	if err != nil {
		fmt.Printf("Error: %v", err)
		c.JSON(http.StatusInternalServerError, err)
	}
	if len(cards) == 0 {
		fmt.Printf("Error: No cards parsed")
		c.JSON(http.StatusInternalServerError, fmt.Errorf("No cards parsed"))
	}

	c.JSON(http.StatusOK, nil)
}

func testHandleImport(t *testing.T, content string) {
	server := server.Configure()
	testSetup(server)
	ts := httptest.NewServer(server)
	defer ts.Close()

	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	var fw io.Writer
	var err error
	if fw, err = w.CreateFormFile("file", "data.csv"); err != nil {
		return
	}

	io.WriteString(fw, content)
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/parse-test", ts.URL), &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	client := ts.Client()
	res, err := client.Do(req)
	if err != nil {
		return
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}
