# Frontend

Use this image to run the frontend.

NOTE - If connecting to the backend from outside the docker network, set the `SSR_API_ADDRESS` environment variable to the backend's address (e.g. `SSR_API_ADDRESS="http://host.docker.internal:8000`)
docker run -d \
-p 3000:3000 \
-e SSR_API_ADDRESS="http://host.docker.internal:8000" \
-e PUBLIC_API_ADDRESS="http://localhost:8000" \
storyden-web