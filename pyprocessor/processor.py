from abc import ABC, abstractmethod
from brain_context import BrainContext

class Processor(ABC):
    @abstractmethod
    def process(self, ctx: BrainContext):
        pass

