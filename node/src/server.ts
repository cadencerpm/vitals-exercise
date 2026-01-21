import { buildServer } from "./app";

const { app } = buildServer();

const addr = process.env.ADDR;
const portEnv = process.env.PORT;

const { host, port } = addr
  ? parseAddr(addr)
  : { host: "0.0.0.0", port: parsePort(portEnv) };

try {
  await app.listen({ host, port });
} catch (error) {
  app.log.error(error, "failed to start server");
  process.exit(1);
}

function parsePort(value: string | undefined): number {
  if (!value) {
    return 3000;
  }
  const port = Number(value);
  if (!Number.isFinite(port) || port <= 0) {
    throw new Error(`invalid PORT: ${value}`);
  }
  return port;
}

function parseAddr(value: string): { host: string; port: number } {
  const trimmed = value.trim();
  const index = trimmed.lastIndexOf(":");
  if (index === -1) {
    return { host: "0.0.0.0", port: parsePort(trimmed) };
  }
  const host = trimmed.slice(0, index) || "0.0.0.0";
  const portPart = trimmed.slice(index + 1);
  return { host, port: parsePort(portPart) };
}
