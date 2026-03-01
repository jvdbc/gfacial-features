#/usr/bin/env zsh

PATH_TO_IMAGE="$(pwd)/resources/visage.jpg"

curl -X POST http://localhost:8080/upload-face \
     -H "Content-Type: multipart/form-data" \
     -F "file=@$PATH_TO_IMAGE;type=image/jpeg"