FROM calciumion/new-api:latest

COPY patches/webhook/settle_webhook.sh /opt/tokenjoy/settle_webhook.sh
RUN chmod +x /opt/tokenjoy/settle_webhook.sh

ENV TOKENJOY_SETTLE_WEBHOOK=/opt/tokenjoy/settle_webhook.sh

# When upstream does not invoke MANAGEMENT_WEBHOOK_URL natively, use log-poller.sh
# or backend worker compensateLogs as fallback (see apps/newapi/scripts/log-poller.sh).
