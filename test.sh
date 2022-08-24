#/bin/bash
#GOOGLE_PROJECT_ID=tamr-ops-sandbox HTTPURL_PATH="/api/service/version" HTTPURL_PORT=9100 ./build/Darwin/ch -c gcp -f tamr-version -v -o getInstances -a log
clear ; make _build; GOOGLE_PROJECT_ID=tamr-ops-sandbox ./build/Darwin/ch -c gcp -f idle  -v -o getInstances -a json 
