import { useState, useRef, useEffect } from 'react';

export default function ResizablePanels({ leftPanel, rightPanel, defaultLeftWidth = 30 }) {
  const [leftWidth, setLeftWidth] = useState(defaultLeftWidth);
  const [isResizing, setIsResizing] = useState(false);
  const containerRef = useRef(null);

  // Загрузка сохраненной ширины из localStorage
  useEffect(() => {
    const savedWidth = localStorage.getItem('executionPanelLeftWidth');
    if (savedWidth) {
      setLeftWidth(parseFloat(savedWidth));
    }
  }, []);

  const handleMouseDown = (e) => {
    e.preventDefault();
    setIsResizing(true);
  };

  const handleMouseMove = (e) => {
    if (!isResizing || !containerRef.current) return;

    const container = containerRef.current;
    const containerRect = container.getBoundingClientRect();
    const newWidth = ((e.clientX - containerRect.left) / containerRect.width) * 100;

    // Ограничиваем ширину от 20% до 50%
    if (newWidth >= 10 && newWidth <= 50) {
      setLeftWidth(newWidth);
      localStorage.setItem('executionPanelLeftWidth', newWidth.toString());
    }
  };

  const handleMouseUp = () => {
    setIsResizing(false);
  };

  useEffect(() => {
    if (isResizing) {
      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);

      return () => {
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
      };
    }
  }, [isResizing]);

  return (
    <div
      ref={containerRef}
      className={`resizable-panels ${isResizing ? 'resizing' : ''}`}
    >
      <div className="resizable-left" style={{ width: `${leftWidth}%` }}>
        {leftPanel}
      </div>

      <div className="resizer" onMouseDown={handleMouseDown}>
        <div className="resizer-line" />
      </div>

      <div className="resizable-right" style={{ width: `${100 - leftWidth}%` }}>
        {rightPanel}
      </div>
    </div>
  );
}
