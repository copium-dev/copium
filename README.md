**NOTE:** New features are being paused for a bit so we can finalize a merger with [cvrve](https://cvrve.me)! This insane three data store infra will be migrated all to Postgres. From testing, pagination is faster, analytics are recalculated *way* faster, and search has not been affected; we'd argue it's even a bit better.

If you'd like to see the current progress of things, check out the [cvrve branch](https://github.com/copium-dev/copium/tree/cvrve), you'll see a much cleaner codebase there :)

Extra features coming *immediately* after the merger will be:
- A more accurate and faster company logo fetching CDN
- Average time to first response will be split into average time to first positive and average time to first negative response
- Timeline viewing will look much better
- Timeline supports editing event time instead of being only the time you clicked the new status (for truly accurate analytics when you move from a previous tracking solution to ours)
- Job board data is more fresh, and search/filtering on it is faster since we can now directly integrate with [cvrve](https://cvrve.me) tooling
- And possibly more!
 
# copium.dev
- **frontend:** SvelteKit & Vercel
- **search engine:** Algolia
- **rest api & storage:** Go & Firestore
- **messaging service:** Google Cloud Pub/Sub
- **real-time analytics:** BigQuery & CQRS architecture
- **deployment:** GCP & Docker/Docker-Compose & Traefik
- **job board scraping:** Python

### architectural decisions:
- **why pub/sub?:** previously was using RabbitMQ but we wanted more features (that consume from the same data) so for one-to-many messaging we made a switch to pub/sub
  - **push or pull-based?:** in development we use a pull-based model, in production we use a push-based model. this is mainly to leverage the 2m requests/month free tier of Cloud Run
  - **how are you staying consistent?:** since consumers ack on message processing completion which forces pub/sub to retry, we use compensating transactions: if message publish fails, then rollback database change. else, we can be confident that the message will eventually be processed
- **why CQRS?:** analytic queries could take a while so they should be calculated at write-time, also this keeps us in the 10tb data scanning free tier of BigQuery
  - **wait, why OLAP DBMS?:** it is true that a data warehouse like BigQuery is not optimized for high write volumes, and we are recalculating analytics every time a user updates an application, i.e. we must write in addition to the query. but the analytics queries require a lot of aggregations... just look at `bigquery-consumer/job/job.go`. this tradeoff is worth it due to the complexity of these queries
  - **ok... but what about something like ClickHouse?:** it's expensive. thats it
  - **final question... why not Kafka?:** Kafka is not optimized for the type of queries needed in our analytics, and is also used to process events as they flow in, not on historical data
- **why vercel?:** original plan was to use Firebase, but Svelte 5 was hard to deploy on Firebase and Vercel is just very convenient
- **why algolia?:** no credit card required free plan
- **why go?:** with a serverless model, quick cold starts are great. also, Algolia and BigQuery consumers do potentially long background processing, so the strong concurrency model is critical for scale
  - **why cloud run?:** cloud run is different from the traditional serverless model; each instance can handle many concurrent requests rather than serving only one user at a time. this pairs great with go's http server implementation that, by default, serves requests concurrently
- **why firestore?:** speed is of upmost importance... it also has a free tier
- **why traefik?:** automatically handles SSL certification renewal which Nginx doesn't natively handle and does not support hot renewal with new certificates

![image](https://github.com/user-attachments/assets/4f9655e1-a821-4c7f-ad0c-d3421bcedc1b)

