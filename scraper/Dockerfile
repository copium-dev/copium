# build stage
FROM python:3.12-alpine AS builder
WORKDIR /app

# copy and install requirements, no cache just in case
# then, create wheels for requirements to use in final image
# we also remove the cache to keep image small
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt && \
    pip wheel --no-cache-dir --no-deps --wheel-dir /app/wheels -r requirements.txt && \
    rm -rf /root/.cache/pip/*

# final stage; slim python image with only the built wheels
FROM python:3.12-alpine
WORKDIR /app

# install git cuz we need it for scraping
RUN apk add --no-cache git

# copy wheels from builder (/app/wheels) to final image (/wheels)
# then, run the wheels
COPY --from=builder /app/wheels /wheels
RUN pip install --no-cache /wheels/*

# finally, copy only main.py (for smaller image) and set the environment
# load env variables inside the VM, don't copy into the build
# don't expose this port. this does not serve traffic
COPY main.py .
ENV ENVIRONMENT=prod
CMD ["python", "main.py"]