import { setActivePinia, createPinia } from 'pinia';
import { useInboxStore } from '../inbox';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import * as api from '@/utils/api';

// Mock the api module
vi.mock('@/utils/api', () => ({
  getInboxEmails: vi.fn(),
}));

describe('Pinia Store: inbox', () => {
  beforeEach(() => {
    // creates a fresh pinia and make it active so it's automatically picked
    // up by any useStore() call without having to pass it to it:
    // `useStore(pinia)`
    setActivePinia(createPinia());
  });

  it('initializes with correct default values', () => {
    const store = useInboxStore();
    expect(store.emails).toEqual([]);
    expect(store.isLoading).toBe(false);
    expect(store.error).toBe(null);
  });

  it('fetchEmails successfully fetches and stores emails', async () => {
    const store = useInboxStore();
    const mockEmails = [{ id: 1, subject: 'Test' }];
    api.getInboxEmails.mockResolvedValue({ data: mockEmails });

    await store.fetchEmails();

    expect(store.isLoading).toBe(false);
    expect(store.emails).toEqual(mockEmails);
    expect(store.error).toBe(null);
    expect(api.getInboxEmails).toHaveBeenCalledTimes(1);
  });

  it('fetchEmails handles API errors correctly', async () => {
    const store = useInboxStore();
    api.getInboxEmails.mockRejectedValue(new Error('API Error'));

    await store.fetchEmails();

    expect(store.isLoading).toBe(false);
    expect(store.emails).toEqual([]);
    expect(store.error).toBe('Failed to fetch emails.');
  });

  it('getEmailById returns the correct email', () => {
    const store = useInboxStore();
    store.emails = [
      { id: 1, subject: 'Email 1' },
      { id: 2, subject: 'Email 2' },
    ];

    const email = store.getEmailById('2');
    expect(email).toEqual({ id: 2, subject: 'Email 2' });
  });

  it('getEmailById returns undefined for non-existent email', () => {
    const store = useInboxStore();
    store.emails = [
      { id: 1, subject: 'Email 1' },
    ];

    const email = store.getEmailById('99');
    expect(email).toBeUndefined();
  });
});