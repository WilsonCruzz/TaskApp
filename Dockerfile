# 使用 Node.js 官方镜像来构建 React 应用
FROM node:18 AS build

# 设置工作目录
WORKDIR /app

# 复制前端项目文件
COPY ./client/package.json ./client/
RUN npm install --legacy-peer-deps
COPY ./client ./client/
RUN npm run build

# 使用 Go 镜像来构建 Go 后端
FROM golang:1.18 AS go-build

WORKDIR /app

# 复制 Go 源码文件
COPY ./go.mod ./go.sum ./
RUN go mod tidy
COPY . .

# 构建 Go 应用
RUN go build -o TaskApp .

# 最终镜像，合并 React 构建和 Go 后端
FROM nginx:alpine

# 复制前端构建的文件到 Nginx 目录
COPY --from=build /app/client/build /usr/share/nginx/html

# 复制 Go 应用到最终镜像
COPY --from=go-build /app/TaskApp /usr/bin/TaskApp

# 暴露端口
EXPOSE 80

# 启动 Go 应用
CMD ["/usr/bin/TaskApp"]