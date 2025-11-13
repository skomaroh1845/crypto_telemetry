#!/usr/bin/env python3
"""
Mock Telemetry Service
Generates traces, metrics, and logs for testing the observability stack
"""

import os
import time
import random
import logging
from typing import Dict, List

from opentelemetry import trace, metrics
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.exporter.otlp.proto.http.metric_exporter import OTLPMetricExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry._logs import set_logger_provider
from opentelemetry.sdk._logs import LoggerProvider, LoggingHandler
from opentelemetry.sdk._logs.export import BatchLogRecordProcessor
from opentelemetry.exporter.otlp.proto.http._log_exporter import OTLPLogExporter
from opentelemetry.metrics import Observation, CallbackOptions


# Configuration
OTEL_ENDPOINT = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-collector:4318")
SERVICE_NAME = os.getenv("OTEL_SERVICE_NAME", "mock-telemetry-service")

# Setup resource attributes
resource = Resource.create({
    "service.name": SERVICE_NAME,
    "service.version": "1.0.0",
    "deployment.environment": "development",
})

# Setup Tracing
trace_provider = TracerProvider(resource=resource)
otlp_trace_exporter = OTLPSpanExporter(
    endpoint=f"{OTEL_ENDPOINT}/v1/traces"
)
trace_provider.add_span_processor(BatchSpanProcessor(otlp_trace_exporter))
trace.set_tracer_provider(trace_provider)
tracer = trace.get_tracer(__name__)

# Setup Metrics
metric_reader = PeriodicExportingMetricReader(
    OTLPMetricExporter(endpoint=f"{OTEL_ENDPOINT}/v1/metrics"),
    export_interval_millis=10000,
)
meter_provider = MeterProvider(resource=resource, metric_readers=[metric_reader])
metrics.set_meter_provider(meter_provider)
meter = metrics.get_meter(__name__)

# Setup Logging
logger_provider = LoggerProvider(resource=resource)
set_logger_provider(logger_provider)
otlp_log_exporter = OTLPLogExporter(endpoint=f"{OTEL_ENDPOINT}/v1/logs")
logger_provider.add_log_record_processor(BatchLogRecordProcessor(otlp_log_exporter))
handler = LoggingHandler(level=logging.INFO, logger_provider=logger_provider)

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)
logger.addHandler(handler)

# Create metrics
request_counter = meter.create_counter(
    name="crypto.requests.total",
    description="Total number of crypto requests",
    unit="1",
)
current_price = 0.0
def price_callback(options: CallbackOptions):
    # Return an iterable of Observation
    yield Observation(value=current_price, attributes={"currency": "USD"})

price_gauge = meter.create_observable_gauge(
    "crypto_price",
    callbacks=[price_callback],
    description="Current cryptocurrency price",
)

latency_histogram = meter.create_histogram(
    name="crypto.request.duration",
    description="Request duration in milliseconds",
    unit="ms",
)

# Mock data
CRYPTOCURRENCIES = ["BTC", "ETH", "USDT", "BNB", "SOL", "XRP", "ADA", "DOGE"]
EXCHANGES = ["binance", "coinbase", "kraken", "bybit", "okx"]
OPERATIONS = ["fetch_price", "fetch_orderbook", "fetch_trades", "fetch_ticker"]
USER_AGENTS = ["api-client/1.0", "api-client/2.0", "web-client/1.5"]


def simulate_crypto_operation(operation: str, symbol: str, exchange: str):
    """Simulate a crypto trading operation with traces, metrics, and logs"""
    
    with tracer.start_as_current_span(
        operation,
        attributes={
            "crypto.symbol": symbol,
            "crypto.exchange": exchange,
            "operation.type": operation,
        }
    ) as span:
        try:
            # Simulate processing time
            processing_time = random.uniform(10, 500)
            time.sleep(processing_time / 1000.0)  # Convert to seconds
            
            # Generate random price
            base_prices = {
                "BTC": 45000, "ETH": 2500, "USDT": 1,
                "BNB": 300, "SOL": 100, "XRP": 0.6,
                "ADA": 0.5, "DOGE": 0.08
            }
            price = base_prices.get(symbol, 100) * random.uniform(0.95, 1.05)
            
            # Record metrics
            request_counter.add(
                1,
                attributes={
                    "symbol": symbol,
                    "exchange": exchange,
                    "operation": operation,
                    "status": "success"
                }
            )
            
            latency_histogram.record(
                processing_time,
                attributes={
                    "symbol": symbol,
                    "exchange": exchange,
                    "operation": operation,
                }
            )
            
            # Simulate occasional errors (10% chance)
            if random.random() < 0.1:
                error_msg = random.choice([
                    "Rate limit exceeded",
                    "Connection timeout",
                    "Invalid API key",
                    "Market closed"
                ])
                
                span.set_attribute("error", True)
                span.set_attribute("error.message", error_msg)
                
                logger.error(
                    f"Error in {operation} for {symbol} on {exchange}: {error_msg}",
                    extra={
                        "symbol": symbol,
                        "exchange": exchange,
                        "operation": operation,
                        "error": error_msg,
                    }
                )
                
                request_counter.add(
                    1,
                    attributes={
                        "symbol": symbol,
                        "exchange": exchange,
                        "operation": operation,
                        "status": "error"
                    }
                )
                return None
            
            # Success case
            span.set_attribute("crypto.price", price)
            span.set_attribute("http.status_code", 200)
            
            logger.info(
                f"Successfully executed {operation} for {symbol} on {exchange}: ${price:.2f}",
                extra={
                    "symbol": symbol,
                    "exchange": exchange,
                    "operation": operation,
                    "price": price,
                    "duration_ms": processing_time,
                }
            )
            
            return price
            
        except Exception as e:
            span.set_attribute("error", True)
            span.record_exception(e)
            logger.exception(
                f"Unexpected error in {operation}",
                extra={
                    "symbol": symbol,
                    "exchange": exchange,
                    "operation": operation,
                }
            )
            raise


def simulate_complex_workflow():
    """Simulate a complex workflow with nested spans"""
    
    with tracer.start_as_current_span("process_trading_signal") as parent_span:
        symbol = random.choice(CRYPTOCURRENCIES)
        exchange = random.choice(EXCHANGES)
        
        parent_span.set_attribute("crypto.symbol", symbol)
        parent_span.set_attribute("crypto.exchange", exchange)
        parent_span.set_attribute("workflow.type", "trading_signal")
        
        logger.info(
            f"Starting trading signal processing for {symbol} on {exchange}",
            extra={"symbol": symbol, "exchange": exchange}
        )
        
        # Step 1: Fetch current price
        with tracer.start_as_current_span("fetch_current_price"):
            price = simulate_crypto_operation("fetch_price", symbol, exchange)
            if price:
                parent_span.set_attribute("crypto.current_price", price)
        
        # Step 2: Analyze orderbook
        with tracer.start_as_current_span("analyze_orderbook"):
            simulate_crypto_operation("fetch_orderbook", symbol, exchange)
            time.sleep(random.uniform(0.05, 0.15))
        
        # Step 3: Check recent trades
        with tracer.start_as_current_span("check_recent_trades"):
            simulate_crypto_operation("fetch_trades", symbol, exchange)
            time.sleep(random.uniform(0.03, 0.1))
        
        # Step 4: Make decision
        with tracer.start_as_current_span("make_trading_decision") as decision_span:
            decision = random.choice(["BUY", "SELL", "HOLD"])
            decision_span.set_attribute("trading.decision", decision)
            decision_span.set_attribute("trading.confidence", random.uniform(0.6, 0.95))
            
            logger.info(
                f"Trading decision: {decision} for {symbol}",
                extra={
                    "symbol": symbol,
                    "exchange": exchange,
                    "decision": decision,
                    "price": price,
                }
            )
        
        parent_span.set_attribute("workflow.status", "completed")
        logger.info(
            f"Completed trading signal processing for {symbol}",
            extra={"symbol": symbol, "exchange": exchange}
        )


def main():
    """Main loop to continuously generate telemetry data"""
    logger.info(f"Starting {SERVICE_NAME}")
    logger.info(f"Sending telemetry to {OTEL_ENDPOINT}")
    
    iteration = 0
    
    try:
        while True:
            iteration += 1
            logger.info(f"Starting iteration {iteration}")
            
            # Generate simple operations
            for _ in range(random.randint(3, 8)):
                symbol = random.choice(CRYPTOCURRENCIES)
                exchange = random.choice(EXCHANGES)
                operation = random.choice(OPERATIONS)
                
                simulate_crypto_operation(operation, symbol, exchange)
                time.sleep(random.uniform(0.1, 0.5))
            
            # Generate complex workflows
            for _ in range(random.randint(1, 3)):
                simulate_complex_workflow()
                time.sleep(random.uniform(0.5, 1.0))
            
            logger.info(f"Completed iteration {iteration}")
            
            # Wait before next iteration
            sleep_time = random.uniform(2, 5)
            logger.debug(f"Sleeping for {sleep_time:.2f} seconds")
            time.sleep(sleep_time)
            
    except KeyboardInterrupt:
        logger.info("Shutting down mock telemetry service")
    except Exception as e:
        logger.exception("Fatal error in main loop")
        raise


if __name__ == "__main__":
    main()

