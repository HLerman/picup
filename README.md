Provide a minimalist upload system.
Can be used for both upload and download.

# Dependencies

* github.com/spf13/viper
* github.com/h2non/filetype

# Usage

``` Bash
$ git clone https://github.com/hlerman/picup.git
$ go build
$ chmod +x picup
./picup
```

picup is now running, simply try to send a picture

``` Bash
$ curl -F 'file=@path/to/local/file' http://127.0.0.1:8090/upload
```

# Config

Configuration file can be in JSON, TOML or YAML in the current directory; picup will lookup for .config.json, .config.toml or .config

| entry            | value            | comment          |
| ---------------- | ---------------- | ---------------- |
| port             | string           | server port      |
| baseUrl          | string           | base url, useful for proxy |
| directory        | string           | directory used to store upladed file |
| maxSizeInMB      | int64            | maximum upload size in MB |
| acceptedFileType | string array     | array of mime accepted type, complete available list : https://github.com/h2non/filetype#supported-types |

Complete url is formed with baseUrl, directory and file path.
Ex : (http://127.0.0.1:8090/) + (download/) + (zaefgrd/file)
