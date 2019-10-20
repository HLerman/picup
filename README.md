Provide a minimalist upload system.
Can be used for both upload and download.

#Usage

``` Bash
$ picup -d ./download -p 8090
curl -F 'file=@path/to/local/file' http://127.0.0.1:8090/upload
```

