package swagger

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/jiandahao/goutils/convjson"
	"github.com/jiandahao/goutils/files"
	swaggerFiles "github.com/swaggo/files"

	"github.com/gin-gonic/gin"
)

// Config stores swagger configuration variables.
type Config struct {
	//The url pointing to API definition (normally point to a file with ".json" or ".yaml" suffix).
	URL         string
	URLs        []URLInfo // The urls pointing to API definitions
	folder      string    // where the API definition files are stored
	prefix      string    // prefix of URI for requesting API definition files
	DeepLinking bool
}

// URLInfo url info
type URLInfo struct {
	URL     string
	Summary string
}

// LoadFiles loads swagger api definitions under Folder folder, and joins prefix and
// file path as the final URI for api definitions
func LoadFiles(folder string, prefix string) *Config {
	filePaths, err := files.GetAllFiles(folder, func(filename string) bool {
		return strings.HasSuffix(filename, "swagger.json")
	})

	if err != nil {
		return nil
	}

	var urls []URLInfo
	for _, filePath := range filePaths {
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			panic(err)
		}

		sepc := map[string]interface{}{}
		json.Unmarshal(data, &sepc)
		v := convjson.NewValue(sepc)

		value, err := v.Get("info.title")
		if err != nil {
			continue
		}

		title, err := value.String()
		if err != nil {
			continue
		}

		urls = append(urls, URLInfo{
			Summary: fmt.Sprintf("%s", title),
			URL:     prefix + strings.TrimPrefix(filePath, folder),
		})
	}

	c := &Config{}

	c.URLs = append(c.URLs, urls...)
	c.folder = folder
	c.prefix = prefix
	return c
}

// HTTPSwaggerHandle serves files from the given file system root, and serves swagger relative files.
//
// Path "/swagger_definitions" is used to serve static resource.
//func HTTPSwaggerHandle(r *gin.Engine, relativePath string, root string) {
//	r.Static("/swagger_definitions", root)
//	relativePath = strings.TrimSuffix(relativePath, "/")
//	if !strings.HasSuffix(relativePath, "/*any") {
//		relativePath = relativePath + "/*any"
//	}
//	r.GET(relativePath, WrapHandler(swaggerFiles.Handler, URLsFrom(root, "/swagger_definitions")))
//}

// GinHandler wrapper for gin
func GinHandler(cfg *Config) gin.HandlerFunc {
	serveSwaggerFileHandler := HTTPHandler(cfg)
	return func(c *gin.Context) {
		serveSwaggerFileHandler(c.Writer, c.Request)
	}
}

// HTTPHandler serve swagger files wrapper
func HTTPHandler(config *Config) http.HandlerFunc {
	type swaggerUIBundle struct {
		URL         string
		URLs        []URLInfo
		DeepLinking bool
	}

	var rexp = regexp.MustCompile(`(.*)(index\.html|.*\.json|favicon-16x16\.png|favicon-32x32\.png|/oauth2-redirect\.html|swagger-ui\.css|swagger-ui\.css\.map|swagger-ui\.js|swagger-ui\.js\.map|swagger-ui-bundle\.js|swagger-ui-bundle\.js\.map|swagger-ui-standalone-preset\.js|swagger-ui-standalone-preset\.js\.map)[\?|.]*`)

	t := template.New("swagger_index.html")
	index, _ := t.Parse(swaggerIndexTempl)

	h := swaggerFiles.Handler

	return func(w http.ResponseWriter, r *http.Request) {
		var matches []string
		if matches = rexp.FindStringSubmatch(r.RequestURI); len(matches) != 3 {
			w.WriteHeader(404)
			w.Write([]byte("404 page not found"))
			return
		}

		path := matches[2]
		prefix := matches[1]
		h.Prefix = prefix

		if strings.HasSuffix(path, ".html") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		} else if strings.HasSuffix(path, ".css") {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		} else if strings.HasSuffix(path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(path, ".json") {
			w.Header().Set("Content-Type", "application/json")
		}

		switch path {
		case "index.html":
			if err := index.Execute(w, &swaggerUIBundle{
				URL:         config.URL,
				URLs:        config.URLs,
				DeepLinking: config.DeepLinking,
			}); err != nil {
				w.Write([]byte(err.Error()))
			}
		default:
			if strings.HasSuffix(path, ".json") {
				filePathInFileSystem := strings.Replace(r.RequestURI, config.prefix, config.folder, 1)
				_, err := os.Stat(filePathInFileSystem)
				if os.IsNotExist(err) {
					w.WriteHeader(404)
					w.Write([]byte("404 page not found"))
					return
				}

				fileByte, err := ioutil.ReadFile(filePathInFileSystem)
				if err != nil {
					w.WriteHeader(404)
					w.Write([]byte("404 page not found"))
					return
				}

				w.Write(fileByte)
				return
			}
			h.ServeHTTP(w, r)
		}
	}
}

const swaggerIndexTempl = `<!-- HTML for static distribution bundle build -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Swagger UI</title>
  <link href="https://fonts.googleapis.com/css?family=Open+Sans:400,700|Source+Code+Pro:300,600|Titillium+Web:400,600,700" rel="stylesheet">
  <link rel="stylesheet" type="text/css" href="./swagger-ui.css" >
  <link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32" />
  <link rel="icon" type="image/png" href="./favicon-16x16.png" sizes="16x16" />
  <style>
    html
    {
        box-sizing: border-box;
        overflow: -moz-scrollbars-vertical;
        overflow-y: scroll;
    }
    *,
    *:before,
    *:after
    {
        box-sizing: inherit;
    }

    body {
      margin:0;
      background: #fafafa;
    }
  </style>
</head>

<body>

<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" style="position:absolute;width:0;height:0">
  <defs>
    <symbol viewBox="0 0 20 20" id="unlocked">
          <path d="M15.8 8H14V5.6C14 2.703 12.665 1 10 1 7.334 1 6 2.703 6 5.6V6h2v-.801C8 3.754 8.797 3 10 3c1.203 0 2 .754 2 2.199V8H4c-.553 0-1 .646-1 1.199V17c0 .549.428 1.139.951 1.307l1.197.387C5.672 18.861 6.55 19 7.1 19h5.8c.549 0 1.428-.139 1.951-.307l1.196-.387c.524-.167.953-.757.953-1.306V9.199C17 8.646 16.352 8 15.8 8z"></path>
    </symbol>

    <symbol viewBox="0 0 20 20" id="locked">
      <path d="M15.8 8H14V5.6C14 2.703 12.665 1 10 1 7.334 1 6 2.703 6 5.6V8H4c-.553 0-1 .646-1 1.199V17c0 .549.428 1.139.951 1.307l1.197.387C5.672 18.861 6.55 19 7.1 19h5.8c.549 0 1.428-.139 1.951-.307l1.196-.387c.524-.167.953-.757.953-1.306V9.199C17 8.646 16.352 8 15.8 8zM12 8H8V5.199C8 3.754 8.797 3 10 3c1.203 0 2 .754 2 2.199V8z"/>
    </symbol>

    <symbol viewBox="0 0 20 20" id="close">
      <path d="M14.348 14.849c-.469.469-1.229.469-1.697 0L10 11.819l-2.651 3.029c-.469.469-1.229.469-1.697 0-.469-.469-.469-1.229 0-1.697l2.758-3.15-2.759-3.152c-.469-.469-.469-1.228 0-1.697.469-.469 1.228-.469 1.697 0L10 8.183l2.651-3.031c.469-.469 1.228-.469 1.697 0 .469.469.469 1.229 0 1.697l-2.758 3.152 2.758 3.15c.469.469.469 1.229 0 1.698z"/>
    </symbol>

    <symbol viewBox="0 0 20 20" id="large-arrow">
      <path d="M13.25 10L6.109 2.58c-.268-.27-.268-.707 0-.979.268-.27.701-.27.969 0l7.83 7.908c.268.271.268.709 0 .979l-7.83 7.908c-.268.271-.701.27-.969 0-.268-.269-.268-.707 0-.979L13.25 10z"/>
    </symbol>

    <symbol viewBox="0 0 20 20" id="large-arrow-down">
      <path d="M17.418 6.109c.272-.268.709-.268.979 0s.271.701 0 .969l-7.908 7.83c-.27.268-.707.268-.979 0l-7.908-7.83c-.27-.268-.27-.701 0-.969.271-.268.709-.268.979 0L10 13.25l7.418-7.141z"/>
    </symbol>


    <symbol viewBox="0 0 24 24" id="jump-to">
      <path d="M19 7v4H5.83l3.58-3.59L8 6l-6 6 6 6 1.41-1.41L5.83 13H21V7z"/>
    </symbol>

    <symbol viewBox="0 0 24 24" id="expand">
      <path d="M10 18h4v-2h-4v2zM3 6v2h18V6H3zm3 7h12v-2H6v2z"/>
    </symbol>

  </defs>
</svg>

<div id="swagger-ui"></div>

<script src="./swagger-ui-bundle.js"> </script>
<script src="./swagger-ui-standalone-preset.js"> </script>
<script>
window.onload = function() {
  // Build a system
  const ui = SwaggerUIBundle({
	url: "{{.URL}}",
    urls: [
      {{ range .URLs }}
      {
          name:"{{ .Summary }}",
          url:" {{ .URL }}"
      },
      {{- end }}
    ],
    dom_id: '#swagger-ui',
    validatorUrl: null,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
	layout: "StandaloneLayout",
	deepLinking: {{.DeepLinking}}
  })

  window.ui = ui
}
</script>
</body>

</html>
`
