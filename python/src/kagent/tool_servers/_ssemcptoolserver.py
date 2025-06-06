from autogen_core import Component
from autogen_ext.tools.mcp._config import SseServerParams
from autogen_ext.tools.mcp._factory import mcp_server_tools
from loguru import logger

from ._tool_server import ToolServer


class SseMcpToolServerConfig(SseServerParams):
    pass


class SseMcpToolServer(ToolServer, Component[SseMcpToolServerConfig]):
    component_config_schema = SseMcpToolServerConfig
    component_type = "tool_server"
    component_provider_override = "kagent.tool_servers.SseMcpToolServer"

    def __init__(self, config: SseMcpToolServerConfig):
        self.config = config

    async def discover_tools(self) -> list[Component]:
        try:
            logger.debug(f"Discovering tools from sse server: {self.config}")
            tools = await mcp_server_tools(self.config)
            return tools
        except Exception as e:
            raise Exception(f"Failed to discover tools: {e}") from e

    def _to_config(self) -> SseMcpToolServerConfig:
        return SseMcpToolServerConfig(**self.config.model_dump())

    @classmethod
    def _from_config(cls, config: SseMcpToolServerConfig):
        return cls(config)
