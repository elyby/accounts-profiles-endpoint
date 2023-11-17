# Profile endpoint for Account Ely.by

This is a "microservice" which solves PHP's inability to handle large numbers of small requests.

We faced the problem that the [endpoint for accessing player's profile by uuid](https://docs.ely.by/en/minecraft-auth.html#profile-request) become receiving a huge number of requests. For PHP, initialization of the framework, database connections and so on is more expensive than the query processing itself. So I wrote this endpoint in Go to take the load off PHP.

This project is not an example of a perfect architecture or anything else. It just does its job by being written very quickly. If I have the mood and, more importantly, the time, I will tidy up this code. But here and now it does its job and that is the most important thing.
