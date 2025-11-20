import React from 'react';

// Регулярное выражение для поиска URL
const URL_REGEX = /(https?:\/\/[^\s]+)/g;
// Регулярное выражение для поиска Telegram ссылок
const TELEGRAM_REGEX = /(@[a-zA-Z0-9_]+|t\.me\/[a-zA-Z0-9_]+)/g;

interface LinkifyProps {
  text: string;
  className?: string;
}

/**
 * Компонент для преобразования текста с URL в кликабельные ссылки
 */
export function Linkify({ text, className = '' }: LinkifyProps) {
  if (!text) return null;

  const parts: (string | JSX.Element)[] = [];
  let lastIndex = 0;
  let key = 0;

  // Сначала обрабатываем URL
  let match;
  const urlMatches: Array<{ start: number; end: number; url: string }> = [];

  while ((match = URL_REGEX.exec(text)) !== null) {
    urlMatches.push({
      start: match.index,
      end: match.index + match[0].length,
      url: match[0],
    });
  }

  // Затем обрабатываем Telegram ссылки (если они не являются частью URL)
  const telegramMatches: Array<{ start: number; end: number; text: string; url: string }> = [];
  URL_REGEX.lastIndex = 0; // Сбрасываем regex

  while ((match = TELEGRAM_REGEX.exec(text)) !== null) {
    // Проверяем, не является ли это частью уже найденного URL
    const isPartOfUrl = urlMatches.some(
      (urlMatch) => match.index >= urlMatch.start && match.index < urlMatch.end
    );

    if (!isPartOfUrl) {
      const telegramText = match[0];
      const telegramUrl =
        telegramText.startsWith('@')
          ? `https://t.me/${telegramText.slice(1)}`
          : `https://${telegramText}`;

      telegramMatches.push({
        start: match.index,
        end: match.index + telegramText.length,
        text: telegramText,
        url: telegramUrl,
      });
    }
  }

  // Объединяем все совпадения и сортируем по позиции
  const allMatches = [
    ...urlMatches.map((m) => ({ ...m, type: 'url' as const })),
    ...telegramMatches.map((m) => ({ ...m, type: 'telegram' as const })),
  ].sort((a, b) => a.start - b.start);

  // Удаляем перекрывающиеся совпадения (приоритет URL)
  const filteredMatches: typeof allMatches = [];
  for (const match of allMatches) {
    const overlaps = filteredMatches.some(
      (existing) =>
        (match.start >= existing.start && match.start < existing.end) ||
        (match.end > existing.start && match.end <= existing.end) ||
        (match.start <= existing.start && match.end >= existing.end)
    );

    if (!overlaps) {
      filteredMatches.push(match);
    }
  }

  // Строим массив частей текста
  for (const match of filteredMatches) {
    // Добавляем текст до совпадения
    if (match.start > lastIndex) {
      parts.push(text.substring(lastIndex, match.start));
    }

    // Добавляем ссылку
    const url = match.type === 'url' ? match.url : match.url;
    const linkText = match.type === 'url' ? match.url : match.text;

    parts.push(
      <a
        key={key++}
        href={url}
        target="_blank"
        rel="noopener noreferrer"
        className="text-[#FF0000] hover:underline break-all"
        onClick={(e) => e.stopPropagation()}
      >
        {linkText}
      </a>
    );

    lastIndex = match.end;
  }

  // Добавляем оставшийся текст
  if (lastIndex < text.length) {
    parts.push(text.substring(lastIndex));
  }

  // Если не было совпадений, возвращаем исходный текст
  if (parts.length === 0) {
    return <span className={className}>{text}</span>;
  }

  return <span className={className}>{parts}</span>;
}

