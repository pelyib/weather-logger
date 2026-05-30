#!/bin/bash

FILENAME_SUBFIX=_${GOOS}_${GOARCH}

echo ${FILENAME_SUBFIX}
echo 
echo =======   HTTP   =======
go build -o /app/bin/http${FILENAME_SUBFIX} /app/cmd/http
echo 
echo =======  LOGGER  =======
go build -o /app/bin/logger${FILENAME_SUBFIX} /app/cmd/logger
echo
echo ====== COMMANDER =======
go build -o /app/bin/commander${FILENAME_SUBFIX} /app/cmd/commander
