import { useState } from 'react';
import { Search, Shield } from 'lucide-react';
import { Input } from './ui/input';
import { Button } from './ui/button';

export function BlacklistTab() {
  const [username, setUsername] = useState('');
  const [searchResult, setSearchResult] = useState<'scammer' | 'clean' | null>(null);

  const handleSearch = async () => {
    if (!username.trim()) return;
    
    try {
      const cleanUsername = username.trim().replace('@', '');
      const response = await fetch(`/api/scammer/${cleanUsername}`);
      const data = await response.json();
      
      if (data.safe === false) {
        setSearchResult('scammer');
      } else {
        setSearchResult('clean');
      }
    } catch (error) {
      console.error('Failed to check user:', error);
      setSearchResult('clean'); // Default to safe on error
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearch();
    }
  };

  return (
    <div className="p-4 space-y-6">
      {/* Header */}
      <div className="pt-2">
        <h1 className="text-2xl mb-2">Чёрный список</h1>
        <p className="text-muted-foreground">Проверьте пользователя перед сделкой</p>
      </div>

      {/* Search Field */}
      <div className="space-y-3">
        <div className="relative">
          <Input
            type="text"
            placeholder="@username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            onKeyPress={handleKeyPress}
            className="pr-12 h-12 rounded-xl border-border focus:border-[#FF0000] focus:ring-[#FF0000]"
          />
          <Button
            onClick={handleSearch}
            className="absolute right-1 top-1 h-10 w-10 p-0 bg-[#FF0000] hover:bg-[#CC0000] rounded-lg"
          >
            <Search size={20} />
          </Button>
        </div>
      </div>

      {/* Search Result */}
      {searchResult && (
        <div className="space-y-4 animate-in fade-in duration-300">
          <div
            className={`p-6 rounded-2xl shadow-md border-2 ${
              searchResult === 'scammer'
                ? 'border-red-500 bg-red-50'
                : 'border-green-500 bg-green-50'
            }`}
          >
            <p
              className={`text-center ${
                searchResult === 'scammer' ? 'text-red-700' : 'text-green-700'
              }`}
            >
              {searchResult === 'scammer'
                ? '⚠️ Осторожно! Мошенник'
                : '✅ Юзер не был замечен в мошеннических схемах'}
            </p>
          </div>

          <p className="text-center text-muted-foreground text-sm">
            Если ошибка — обратитесь к менеджеру
          </p>
        </div>
      )}

      {/* Empty State */}
      {!searchResult && (
        <div className="text-center py-12 text-muted-foreground">
          <Shield size={64} className="mx-auto mb-4 opacity-30" />
          <p>Введите username для проверки</p>
        </div>
      )}
    </div>
  );
}
