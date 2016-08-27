package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

type DoWhich int8

const (
	DO_GET DoWhich = iota
	DO_INDEX
	DO_PLAY
)

type VideoFile struct {
	FilePath, FileType string
	Width, Height      int
}

var videoTmpl = `<!DOCTYPE html>
<html lang="en">

<head>
    <title>Video.js | HTML5 Video Player</title>
    <link href="http://vjs.zencdn.net/5.0.2/video-js.css" rel="stylesheet">
    <script src="http://vjs.zencdn.net/ie8/1.1.0/videojs-ie8.min.js"></script>
    <script src="http://vjs.zencdn.net/5.0.2/video.js"></script>
</head>

<body>
  <video id="example_video_1" class="video-js vjs-default-skin" controls preload="none" width="{{.Height}}" height="{{.Width}}" poster="http://vjs.zencdn.net/v/oceans.png" data-setup="{}">
    <source src="{{.FilePath}}" type="{{.FileType}}">
    <p class="vjs-no-js">To view this video please enable JavaScript, and consider upgrading to a web browser that <a href="http://videojs.com/html5-video-support/" target="_blank">supports HTML5 video</a></p>
  </video>
</body>

</html>`

var indexTmpl = `<!DOCTYPE html>
<html lang="en">
<body>
	{{ range $name ,$show := .FileList }}
	<li><a href="{{$name}}{{if $show}}?play=true{{end}}">{{$name}}</a></li>
	{{ end }}
</body>
</html>`

const (
	listenAddress = ":80"
)

var (
	allowTypes = map[string]string{
		"mp4":  "video/mp4",
		"rmvb": "video/rmvb",
		"avi":  "video/avi",
	}

	videoTmp, indexTmp *template.Template
)

func init() {
	if os.Getenv("CK_DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}
	log.SetFormatter(new(log.JSONFormatter))

	var err error
	videoTmp, err = template.New("video").Parse(videoTmpl)
	if err != nil {
		log.Fatalf("new video template faild, error: %s", err.Error())
	}

	indexTmp, err = template.New("index").Parse(indexTmpl)
	if err != nil {
		log.Fatalf("new index template faild, error: %s", err.Error())
	}

}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", route)

	log.Infof("Listenning on %s", listenAddress)
	if err := http.ListenAndServe(listenAddress, mux); err != nil {
		log.Fatal(err)
	}
}

func route(rw http.ResponseWriter, req *http.Request) {
	path, which := parseUrl(rw, req)
	log.WithFields(log.Fields{
		"path":  path,
		"which": which,
	}).Debug("parsed url.")

	var err error
	switch which {
	case DO_INDEX:
		err = index(rw, req, path)
	case DO_GET:
		http.ServeFile(rw, req, path)
	case DO_PLAY:
		err = video(rw, req)
	default:
		err = fmt.Errorf("not support method: %s", which)
	}

	if err != nil {
		log.Errorf("route, ", err.Error())
		rw.WriteHeader(400)
		rw.Write([]byte(err.Error()))
	}
}

func parseUrl(rw http.ResponseWriter, req *http.Request) (path string, which DoWhich) {
	which = DO_INDEX
	if err := req.ParseForm(); err != nil {
		log.Errorf("parseform faild, error: %s", err.Error())
		return
	}

	path = strings.TrimPrefix(req.URL.Path, "/")
	play, _ := strconv.ParseBool(req.FormValue("play"))
	if play {
		which = DO_PLAY
	}
	log.WithFields(log.Fields{
		"path": path,
		"play": play,
	}).Debugf("parsing url: %s", req.URL.String())

	if path == "" {
		return
	}

	fi, err := os.Stat(path)
	if err != nil {
		log.Debugf("find path %s, error: %s", path, err)
		return
	}
	if !fi.IsDir() {
		if !play {
			which = DO_GET
		}
	}
	return
}

func index(rw http.ResponseWriter, req *http.Request, path string) error {
	list, err := getFileList(path)
	if err != nil {
		log.Errorf("list files with path: %s, error: %s", path, err.Error())
		return err
	}

	fileshow := map[string]bool{}
	for _, file := range list {
		if _, ok := allowTypes[getFileSuffix(file)]; ok {
			fileshow[file] = true
			continue
		}
		fileshow[file] = false
	}

	err = indexTmp.Execute(rw, map[string]interface{}{"FileList": fileshow})
	if err != nil {
		return err
	}

	return nil
}

func video(rw http.ResponseWriter, req *http.Request) error {
	file := req.URL.Path
	tagType, ok := allowTypes[getFileSuffix(file)]
	if !ok {
		return fmt.Errorf("not support video file type: %s", tagType)
	}

	err := videoTmp.Execute(rw, &VideoFile{file, tagType, 264, 640})
	if err != nil {
		return err
	}
	return nil
}

func getFileList(path string) ([]string, error) {
	includes := []string{}

	if path == "" {
		path = "."
	}
	err := filepath.Walk(path, func(fpath string, f os.FileInfo, err error) error {
		if f == nil || err != nil {
			return err
		}

		if f.Mode().IsRegular() {
			includes = append(includes, path+"/"+fpath)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return includes, nil
}

func getFileSuffix(file string) string {
	ls := strings.Split(file, ".")
	if len(ls) > 1 {
		return strings.ToLower(ls[len(ls)-1])
	}
	return ""
}
