export default function ContextTab({ steps, selectedStep }) {
  // Находим выбранный шаг или берем последний
  const step = selectedStep
    ? steps.find((s) => s.id === selectedStep)
    : steps[steps.length - 1];

  const context = step?.context || {};

  return (
    <div className="context-viewer">
      {Object.keys(context).length === 0 ? (
        <div className="empty-state">Контекст пуст</div>
      ) : (
        <pre>
          <code>{JSON.stringify(context, null, 2)}</code>
        </pre>
      )}
    </div>
  );
}
