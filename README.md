URL Shortener API
This is a URL shortener API that provides functionalities to shorten URLs, resolve shortened URLs, and get the top domain counts based on shortened URLs.

Endpoints

1. Shorten URL
        POST /shorten
        Description: Shortens a given URL.
        Request Body:
                {
                "url": "https://www.example.com"
                }
        Response:
                {
                "shortenUrl": "http://localhost:4123/resolve/{referenceKey}"
                }
        If the URL is already shortened, it returns the existing shortened URL.

2. Resolve Original URL
        GET /resolve/{referenceKey}
        Description: Retrieves the original URL from the shortened reference key.
        http://localhost:4123/resolve/{referenceKey}

        Response:
            Returns a 301 Moved Permanently HTTP status code and the Location header with the original URL if found.
            If the reference key does not exist, it returns 404 Not Found.

3. Get Top Domains
        GET /domain-counts
        Description: Returns the top domains and their associated count of shortened URLs.

        Response:
                [
                    {
                        "domain": "example.com",
                        "count": 15
                    },
                    {
                        "domain": "example.org",
                        "count": 10
                    }
                ]


Requirements
    Go 1.21.6
    MySQL or MariaDB

Running the API
    make run_url_shortner
