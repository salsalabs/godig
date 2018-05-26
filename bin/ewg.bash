#! /bin/bash
rm -f addressfixer*.log
go run cmd/addressfixer/app/app.go \
--credentials credentials/ewg.yaml \
--criteria 'Email%20IS%20NOT%20EMPTY&condition=Receive_Email>0&condition=State%20IS%20EMPTY&include=supporter_KEY,Email,City,State,Zip,Country,Receive_Email' \
--fixer-count 25 \
--reader-count 25 \
--file-log
