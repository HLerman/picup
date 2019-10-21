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
