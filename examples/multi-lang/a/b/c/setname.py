
from zenmodel import Processor, BrainContext


class SetNameProcessor(Processor):
    def __init__(self, lastname: str):
        self.lastname = lastname
        print(f"SetNameProcessor initialized with firstname: {lastname}")

    def process(self, ctx: BrainContext):
        print("Starting SetNameProcessor.process() method")
        
        name = ctx.get_memory("name")
        name = f"{name} {self.lastname}"
        ctx.set_memory("name", name)

        print(f"Name updated in memory: {name}")
        
        return

