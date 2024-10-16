from abc import ABC, abstractmethod
from brain_context import BrainContextReader

DEFAULT_CAST_GROUP_NAME = "__DEFAULT_CAST_GROUP__"

class Selector(ABC):
    @abstractmethod
    def select(self, ctx: BrainContextReader) -> str:
        pass

