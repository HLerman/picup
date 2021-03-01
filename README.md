Provide a minimalist web upload system.
Can be used for both upload and download.
CORS, Multiple upload and log rotate supported.

# Dependencies

* github.com/spf13/viper
* github.com/h2non/filetype
* github.com/gin-gonic/gin
* github.com/gin-contrib/cors
* github.com/sirupsen/logrus
* gopkg.in/natefinch/lumberjack.v2

# Usage

``` Bash
$ git clone https://github.com/hlerman/picup.git
$ go build
$ chmod +x picup
./picup
```

picup is now running, simply try to send a picture

``` Bash
$ curl -F 'upload=@path/to/local/file' http://127.0.0.1:8090/upload
```

# Config

Configuration file can be in JSON, TOML or YAML in the current directory; picup will lookup for .config.json, .config.toml or .config

| entry            | value            | comment          |
| ---------------- | ---------------- | ---------------- |
| port             | string           | server port      |
| baseUrl          | string           | base url, useful for proxy |
| directory        | string           | directory used to store upladed file |
| virtualDirectory | string           | virtual directory used in url to access to the real directory |
| maxSizeInMB      | int64            | maximum upload size in MB |
| acceptedFileType | string array     | array of mime accepted type, complete available list : https://github.com/h2non/filetype#supported-types |
| mode             | string           | release or debug mode |
| log path         | string           | path of log file |
| log maxSize      | int64            | max size in MB of a log file |
| log maxBackups   | int64            | max backups (log rotate) |
| log maxAge       | int64            | max age of a log file |
| compress         | bool             | compress log file |



Complete url is formed with baseUrl, virtualDirectory and file path.
Ex : (http://127.0.0.1:8090/) + (download/) + (zaefgrd/file)
