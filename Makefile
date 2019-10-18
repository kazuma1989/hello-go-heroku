build: clean
	go build -v .

clean:
	rm -rf ./hello-go-heroku

deploy:
	git push heroku HEAD
	heroku open
