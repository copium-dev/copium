# copium.dev
- **frontend:** SvelteKit & Vercel
- **search engine:** Algolia
- **rest api & storage:** Go & Firestore
- **messaging service:** Google Cloud Pub/Sub
- **real-time analytics:** BigQuery & CQRS architecture
- **deployment:** Docker/Docker-Compose & GCP & Traefik
- **job board scraping** Python

### architectural decisions:
- **why pub/sub?:** previously was using RabbitMQ but we wanted more features (that consume from the same data) so for one-to-many messaging we made a switch to pub/sub
- **why CQRS?:** analytic queries could take a while so they should be calculated at write-time, also this keeps us in the 10tb query transfer data free tier of BigQuery
- **why vercel?:** original plan was to use Firebase, but Svelte 5 was hard to deploy on Firebase and Vercel is just very convenient
- **why algolia?:** no credit card required free plan
- **why go?:** front-facing API is a serverless function, so the quick cold starts were useful. also, our microservices do pretty heavy background processing, so the strong concurrency model was critical for scaling
- **why firestore?:** speed is of upmost importance... it also has a free tier
- **why traefik?:** automatically handles SSL certification renewal which nginx doesn't natively do which gets a bit annoying to set up
