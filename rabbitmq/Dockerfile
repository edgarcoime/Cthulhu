FROM rabbitmq:4-management

# Set container name (this will be overridden by docker-compose)
LABEL container_name="cthulhu-rabbitmq"

# Copy configuration files
COPY rabbitmq.conf /etc/rabbitmq/rabbitmq.conf
# COPY definitions.json /etc/rabbitmq/definitions.json
COPY init.sh /usr/local/bin/init.sh

# Make init script executable
RUN chmod +x /usr/local/bin/init.sh

# Expose ports
EXPOSE 5672 15672 25672

# Use the init script as entrypoint
ENTRYPOINT ["/usr/local/bin/init.sh"]