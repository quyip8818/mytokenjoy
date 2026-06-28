FROM calciumion/new-api:latest

COPY patches/webhook/settle_webhook.sh /opt/tokenjoy/settle_webhook.sh
RUN chmod +x /opt/tokenjoy/settle_webhook.sh

# P0 uses log compensation in Management worker when webhook patch is not applied.
# Optional: mount MANAGEMENT_WEBHOOK_URL env and run sidecar poller from apps/newapi/scripts/log-poller.sh
