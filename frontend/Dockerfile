FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/node:24.14.0-alpine3.22

WORKDIR /app

COPY package.json pnpm-lock.yaml* ./
RUN npm install -g pnpm && pnpm install

COPY . .

EXPOSE 3000

CMD ["pnpm", "dev"]