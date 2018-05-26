#! /bin/bash
rm -f addressfixer*.log
go run cmd/addressfixer/app/app.go \
--credentials credentials/ewg.yaml \
--criteria 'Email%20IS%20NOT%20EMPTY&condition%3DEmail%20LIKE%20&apos;@&apos;.&apos;%40%25.%25&condition%3DReceive_Email>0&condition%3DState%20IS%20EMPTY&include=supporter_KEY,Email,City,State,Zip,Country,Receive_Email' \
--fixer-count 25 \
--reader-count 25 \
--file-log
