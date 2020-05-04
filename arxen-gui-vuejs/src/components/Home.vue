import gql from "graphql-tag";
<template>
    <div class="card-window" :style="[{ height }, cssVars]">
        <div class="chat-container">
            <app-chat-list class="col-2"
                           @selectDiffrentChat="updateSelectedChat"
                           :selectedChatId="selectedChatId"
            />
            <div class="col-10">
                <div v-if="selectedChatId">
                    <div class="row mb-3">
                        <app-message-form :selectedChatId="selectedChatId"/>
                    </div>

                    <div class="row">
                        <app-message-list class="col-12"
                                          :selectedChatId="selectedChatId"
                                          :textMessages="t"
                        />
                    </div>
                </div>
                <div v-if="!selectedChatId">Select Chat</div>
            </div>
        </div>
    </div>
</template>

<script>
    import ChatList from '@/components/ChatList.vue';
    import MessageList from '@/components/MessageList.vue';
    import MessageForm from '@/components/MessageForm.vue';
    import { defaultThemeStyles, cssThemeVars } from '../themes';
    import locales from '../locales'

    export default {
        components: {
            'app-chat-list': ChatList,
            'app-message-list': MessageList,
            'app-message-form': MessageForm,
        },
        props: {
            height: {type: String, default: '600px'},
            theme: { type: String, default: 'light' },
            styles: { type: Object, default: () => ({}) },
        },
        data() {
            return {
                chatList: [],
                selectedChatId: '',
            };
        },
        computed: {
            t() {
                return {
                    ...locales,
                    ...this.textMessages
                }
            },
            cssVars() {
                const defaultStyles = defaultThemeStyles[this.theme];
                const customStyles = {};
                Object.keys(defaultStyles).map(key => {
                    customStyles[key] = {
                        ...defaultStyles[key],
                        ...(this.styles[key] || {})
                    }
                });
                return cssThemeVars(customStyles);
            }
        },
        methods: {
            updateSelectedChat(chatID) {
                this.selectedChatId = chatID;
            }
        }
    };
</script>

<style lang="scss">
    @import '../styles/index.scss';
    * {
        font-family: inherit;
    }
    .card-window {
        width: 100%;
        display: block;
        max-width: 100%;
        background: var(--chat-content-bg-color);
        color: var(--chat-color);
        overflow-wrap: break-word;
        position: relative;
        white-space: normal;
        border: var(--chat-container-border);
        border-radius: var(--chat-container-border-radius);
        box-shadow: var(--chat-container-box-shadow);
    }
        .chat-container {
            height: 100%;
            display: flex;
            textarea,
            input[type='text'],
            input[type='search'] {
                -webkit-appearance: none;
        }
    }

</style>
