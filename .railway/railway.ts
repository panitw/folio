import { defineRailway, project, service } from "railway/iac";

// folio8 Designer: a static site. Railway builds the root Dockerfile (Go 1.26.0 +
// Node 24.16.0 build stage, Caddy serve stage — see deploy/Caddyfile), which
// listens on $PORT.
export default defineRailway(() => {
  const designer = service("folio-designer", {
    healthcheck: "/",
    healthcheckTimeout: 60,
  });

  return project("folio8", {
    resources: [designer],
  });
});
