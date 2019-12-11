# Movie recommender - with data from IMDB
## Crawl data
### Crawl users
* Crawl list of users have many reviews (>=100).
* IMDB do not provide list of users.
* Go through each movie in this link [https://www.imdb.com/chart/top], check all users review on this movie. If the total number of reviews of a user >=100, add to file users.txt.

## Crawl rating
* From user list in users.txt, go to each user page and crawl all rating from this user.