import torch.nn

ATTRIBUTE_NAME = "_nebulgym_backend"


class CallMethodFixingEnvRunner:
    def __init__(self, model: torch.nn.Module):
        self._training_learner = None
        self._model = model

    def __enter__(self):
        self._training_learner = getattr(self._model, ATTRIBUTE_NAME, None)
        if self._training_learner is not None:
            setattr(self._model, ATTRIBUTE_NAME, None)

    def __exit__(self, exc_type, exc_val, exc_tb):
        if self._training_learner is not None:
            setattr(self._model, ATTRIBUTE_NAME, self._training_learner)
