#!/bin/sh

set -x

# load .env
set -o allexport; source .env; set +o allexport
GCP_RUN_SERVICE_ACCOUNT=dl-run@$GCP_ID.iam.gserviceaccount.com

# set GCP project
gcloud config set project $GCP_ID
# build and push container image
docker build -t gcr.io/$GCP_ID/$APP_NAME_APIKEYS:$APP_VERSION -f apikey.Dockerfile .
docker push gcr.io/$GCP_ID/$APP_NAME_APIKEYS:$APP_VERSION

# deploy
# set --max-instances=1 to avoid bitcoin-cli and wallet facing race condition
# env BITCOIN_* use default value so no need to pass them
# env PORT is set by Cloud Run so no need to pass it
gcloud run deploy $APP_NAME_APIKEYS \
  --image gcr.io/$GCP_ID/$APP_NAME_APIKEYS:$APP_VERSION \
  --platform managed \
  --memory=128Mi --cpu=1000m \
  --max-instances=1 \
  --set-env-vars=MONGO_HOSTNAME=$MONGO_HOSTNAME,MONGO_USER=$MONGO_USER,MONGO_PASSWORD=$MONGO_PASSWORD \
  --region=asia-northeast1 \
  --service-account=$GCP_RUN_SERVICE_ACCOUNT \
  --allow-unauthenticated
