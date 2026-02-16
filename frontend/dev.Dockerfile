# Development Dockerfile with Vite dev server for HMR
FROM node:20-alpine

WORKDIR /app

# Copy package files for dependency installation
COPY package.json package-lock.json* ./
RUN npm install

# Source code will be mounted as volume
EXPOSE 8000

CMD ["npm", "run", "dev"]
