FROM node:22-alpine3.19

WORKDIR /uberbase

ADD . .
RUN npm ci
RUN npm install dotenv

ENV NODE_ENV=production
ENV PORT=3000
ENV HOST=0.0.0.0

RUN npm run build

CMD ["node", "build"]
