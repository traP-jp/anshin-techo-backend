FROM swaggerapi/swagger-ui:v5.31.0
COPY docs/openapi/openapi-v1.yaml /docs/openapi-v1.yaml
ENV SWAGGER_JSON=/docs/openapi-v1.yaml