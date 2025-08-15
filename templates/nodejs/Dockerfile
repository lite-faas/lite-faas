FROM node:18-alpine

WORKDIR /app

COPY function.js .
COPY package.json .

RUN npm install --production

EXPOSE 8080

CMD ["node", "function.js"]
