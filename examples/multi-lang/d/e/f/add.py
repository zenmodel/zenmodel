from zenmodel import Processor, BrainContext


class AddProcessor(Processor):
    def __init__(self, a: int, b: int):
        self.a = a
        self.b = b
        print(f"AddProcessor initialized with a: {a}, b: {b}")

    def process(self, ctx: BrainContext):
        print("Starting AddProcessor.process() method")
    
        name = ctx.get_memory("name")
        date = ctx.get_memory("date")
        result = self.a + self.b
        answer = f"hello {name}, today is {date}, {self.a} + {self.b} = {result}"
        ctx.set_memory("answer", answer)
        print(f"Answer updated in memory: {answer}")
        return
