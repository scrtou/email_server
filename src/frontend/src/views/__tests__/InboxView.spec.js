import { mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import InboxView from '../InboxView.vue';
import { useInboxStore } from '@/stores/inbox';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { ElAlert, ElEmpty, ElLoading } from 'element-plus';

// Mock components
const AppLayout = {
  template: '<div><slot/></div>',
};
const EmailListItem = {
  props: ['email'],
  template: '<div class="mock-email-item">{{ email.subject }}</div>',
};

describe('InboxView.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('calls fetchEmails on mount if emails are not already loaded', () => {
    const store = useInboxStore();
    const fetchEmailsSpy = vi.spyOn(store, 'fetchEmails');

    mount(InboxView, {
      global: {
        plugins: [store],
        components: {
          AppLayout,
          EmailListItem,
          ElAlert,
          ElEmpty,
        },
        directives: {
          loading: ElLoading.directive,
        },
      },
    });

    expect(fetchEmailsSpy).toHaveBeenCalledTimes(1);
  });

  it('displays a loading spinner while fetching emails', () => {
    const store = useInboxStore();
    store.isLoading = true;

    const wrapper = mount(InboxView, {
      global: {
        plugins: [store],
        components: {
          AppLayout,
          EmailListItem,
          ElAlert,
          ElEmpty,
        },
        directives: {
          loading: ElLoading.directive,
        },
      },
    });

    expect(wrapper.find('.loading-spinner').exists()).toBe(true);
  });

  it('displays an error message if fetching fails', async () => {
    const store = useInboxStore();
    store.error = 'Failed to load';

    const wrapper = mount(InboxView, {
      global: {
        plugins: [store],
        components: {
          AppLayout,
          EmailListItem,
          ElAlert,
          ElEmpty,
        },
        directives: {
          loading: ElLoading.directive,
        },
      },
    });
    
    expect(wrapper.findComponent(ElAlert).exists()).toBe(true);
    expect(wrapper.findComponent(ElAlert).props('title')).toBe('Failed to load');
  });

  it('displays a list of emails when fetch is successful', () => {
    const store = useInboxStore();
    store.emails = [
      { id: 1, subject: 'Email 1' },
      { id: 2, subject: 'Email 2' },
    ];

    const wrapper = mount(InboxView, {
      global: {
        plugins: [store],
        components: {
          AppLayout,
          EmailListItem,
          ElAlert,
          ElEmpty,
        },
        directives: {
          loading: ElLoading.directive,
        },
      },
    });

    const emailItems = wrapper.findAll('.mock-email-item');
    expect(emailItems.length).toBe(2);
    expect(emailItems[0].text()).toBe('Email 1');
  });

  it('displays an empty state message when there are no emails', () => {
    const store = useInboxStore();
    store.emails = [];
    
    const wrapper = mount(InboxView, {
      global: {
        plugins: [store],
        components: {
          AppLayout,
          EmailListItem,
          ElAlert,
          ElEmpty,
        },
        directives: {
          loading: ElLoading.directive,
        },
      },
    });

    expect(wrapper.findComponent(ElEmpty).exists()).toBe(true);
  });
});