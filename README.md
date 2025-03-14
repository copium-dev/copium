# copium.dev
- **frontend:** SvelteKit & Vercel
- **search engine:** Algolia
- **rest api & storage:** Go & Firestore
- **messaging service:** Google Cloud Pub/Sub
- **real-time analytics:** BigQuery & CQRS architecture
- **deployment:** Docker/Docker-Compose & GCP & Traefik
- **job board scraping:** Python

### architectural decisions:
- **why pub/sub?:** previously was using RabbitMQ but we wanted more features (that consume from the same data) so for one-to-many messaging we made a switch to pub/sub
- **why CQRS?:** analytic queries could take a while so they should be calculated at write-time, also this keeps us in the 10tb query transfer data free tier of BigQuery
  - **wait, why OLAP DBMS?:** it is true that a data warehouse like BigQuery is not optimized for high write volumes, and we are recalculating analytics every time a user updates an application, i.e. we must write in addition to the query. but the analytics queries are kinda crazy, just look at `bigquery-consumer/job/job.go`. this tradeoff is worth it due to the complexity of these queries
  - **ok... but what about something like ClickHouse?:** it's expensive. thats it
- **why vercel?:** original plan was to use Firebase, but Svelte 5 was hard to deploy on Firebase and Vercel is just very convenient
- **why algolia?:** no credit card required free plan
- **why go?:** front-facing API is a serverless function, so the quick cold starts were useful. also, Algolia and BigQuery consumers do potentially long background processing, so the strong concurrency model was critical for scale
- **why firestore?:** speed is of upmost importance... it also has a free tier
- **why traefik?:** automatically handles SSL certification renewal which Nginx doesn't natively handle and does not support hot renewal with new certificates

![image](https://github.com/user-attachments/assets/4f29ede9-6134-49bb-8d55-aa1c597bfde8)

