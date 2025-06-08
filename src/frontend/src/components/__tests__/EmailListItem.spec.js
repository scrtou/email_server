import { mount } from '@vue/test-utils';
import EmailListItem from '../EmailListItem.vue';
import { describe, it, expect, vi } from 'vitest';
import { ElCard } from 'element-plus';

// Mock vue-router
const mockRouter = {
  push: vi.fn(),
};

vi.mock('vue-router', () => ({
  useRouter: () => mockRouter,
}));

describe('EmailListItem.vue', () => {
  const mockEmail = {
    id: 1,
    from: [{ name: 'John Doe', address: 'john.doe@example.com' }],
    subject: 'Hello World',
    snippet: 'This is a test email.',
    date: '2023-10-27T10:00:00Z',
    isRead: false,
  };

  it('renders email information correctly', () => {
    const wrapper = mount(EmailListItem, {
      props: { email: mockEmail },
      global: {
        components: {
          ElCard,
        },
      },
    });

    expect(wrapper.find('.from').text()).toBe('John Doe');
    expect(wrapper.find('.subject').text()).toBe('Hello World');
    expect(wrapper.find('.snippet').text()).toBe('This is a test email.');
    // Check formatted date
    expect(wrapper.find('.date').text()).toBe('2023-10-27 18:00'); // Assuming UTC+8
  });

  it('applies "is-read" class when email is read', () => {
    const readEmail = { ...mockEmail, isRead: true };
    const wrapper = mount(EmailListItem, {
      props: { email: readEmail },
      global: {
        components: {
          ElCard,
        },
      },
    });

    expect(wrapper.find('.from').classes()).toContain('is-read');
    expect(wrapper.find('.subject').classes()).toContain('is-read');
  });

  it('does not apply "is-read" class when email is unread', () => {
    const wrapper = mount(EmailListItem, {
      props: { email: mockEmail },
      global: {
        components: {
          ElCard,
        },
      },
    });

    expect(wrapper.find('.from').classes()).not.toContain('is-read');
    expect(wrapper.find('.subject').classes()).not.toContain('is-read');
  });

  it('navigates to email detail on click', async () => {
    const wrapper = mount(EmailListItem, {
      props: { email: mockEmail },
      global: {
        components: {
          ElCard,
        },
      },
    });

    await wrapper.find('.email-list-item').trigger('click');

    expect(mockRouter.push).toHaveBeenCalledTimes(1);
    expect(mockRouter.push).toHaveBeenCalledWith({
      name: 'EmailDetail',
      params: { id: 1 },
    });
  });
});