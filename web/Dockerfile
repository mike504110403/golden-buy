# 使用 Node.js 官方鏡像
FROM node:22-alpine

# 設置工作目錄
WORKDIR /app

# 複製 package.json 和 pnpm-lock.yaml
COPY package.json pnpm-lock.yaml ./

# 安裝 pnpm
RUN npm install -g pnpm

# 安裝依賴
RUN pnpm install --frozen-lockfile

# 複製源代碼
COPY . .

# 構建應用
RUN pnpm build

# 暴露端口
EXPOSE 3000

# 啟動應用
CMD ["pnpm", "preview", "--host", "0.0.0.0", "--port", "3000"]
