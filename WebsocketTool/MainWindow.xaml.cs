using WebsocketNetTool.Base;
using WebsocketNetTool.WebSocket;
using System.Windows;

namespace WebsocketNetTool
{
    /// <summary>
    /// MainWindow.xaml 
    /// </summary>
    public partial class MainWindow : BaseWindow
    {
        public MainWindow()
        {
            InitializeComponent();
            InitEvents();
        }


     public override void InitEvents()
        {
            BtnWebsocketServer.Click += BtnWebsocketServer_Click;
            BtnWebsocketClient.Click += BtnWebsocketClient_Click;
        }

        private void BtnWebsocketClient_Click(object sender, RoutedEventArgs e)
        {
            new WebsocketClientWindow().Show();
        }

        private void BtnWebsocketServer_Click(object sender, RoutedEventArgs e)
        {
            new WebsocketServerWindow().Show();
        }
    }
}
