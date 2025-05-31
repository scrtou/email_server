# Stage 1: Build the Vue.js application
FROM node:20-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy package.json and package-lock.json (or yarn.lock)
COPY package.json ./
COPY package-lock.json ./

# Install project dependencies
RUN npm install

# Copy the rest of the application code
COPY . .

# Build the application for production
# To inject build-time environment variables, you can use ARG and ENV:
# ARG VUE_APP_API_BASE_URL
# ENV VUE_APP_API_BASE_URL=${VUE_APP_API_BASE_URL}
# Then, when building the image, pass the argument:
# docker build --build-arg VUE_APP_API_BASE_URL=your_api_url -t your-image-name .
RUN npm run build

# Stage 2: Serve the application with Nginx
FROM nginx:alpine

# Copy the build output from the builder stage to Nginx's web root
COPY --from=builder /app/dist /usr/share/nginx/html

# (Optional) Copy a custom Nginx configuration file
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Nginx will start automatically when the container launches