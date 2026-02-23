FROM oven/bun:1

WORKDIR /app

COPY . .

RUN bun install
RUN bun run build

EXPOSE 3000

CMD ["node", "build"]

# Build image: docker build -t protocol-simulator .
# Run container: docker run -p 3000:3000 protocol-simulator
