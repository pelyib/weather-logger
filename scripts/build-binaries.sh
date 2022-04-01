#!/bin/bash

FILENAME_SUBFIX=_${GOOS}_${GOARCH}

go build -o /app/bin/http${FILENAME_SUBFIX} /app/cmd/http
echo 
echo ======= LOGGER =======
printenv
go build -o /app/bin/logger${FILENAME_SUBFIX} /app/cmd/logger
echo
echo ====== COMMANDER =======
printenv
go build -o /app/bin/commander${FILENAME_SUBFIX} /app/cmd/commander
